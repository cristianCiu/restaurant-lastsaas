package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"lastsaas/internal/db"
	"lastsaas/internal/middleware"
	"lastsaas/internal/models"
	"lastsaas/internal/syslog"
	"lastsaas/internal/validation"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

type InventoryHandler struct {
	db     *db.MongoDB
	syslog syslogiface
}

type syslogiface interface {
	LogCat(ctx context.Context, severity models.LogSeverity, category models.LogCategory, message string)
	LogCatWithUser(ctx context.Context, severity models.LogSeverity, category models.LogCategory, message string, userID primitive.ObjectID)
}

func NewInventoryHandler(database *db.MongoDB, logger *syslog.Logger) *InventoryHandler {
	return &InventoryHandler{db: database, syslog: logger}
}

func (h *InventoryHandler) getTenantID(r *http.Request) (primitive.ObjectID, bool) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		return primitive.NilObjectID, false
	}
	return tenant.ID, true
}

func (h *InventoryHandler) getUserID(r *http.Request) (primitive.ObjectID, bool) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		return primitive.NilObjectID, false
	}
	return user.ID, true
}

// ListStockItems returns stock items for a location, grouped by category.
func (h *InventoryHandler) ListStockItems(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	locationIDStr := r.URL.Query().Get("location_id")
	if locationIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	locationID, err := primitive.ObjectIDFromHex(locationIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location_id")
		return
	}

	filter := bson.M{"tenantId": tenantID, "locationId": locationID}
	opts := options.Find().SetSort(bson.D{{Key: "category", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := h.db.StockItems().Find(r.Context(), filter, opts)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch stock items")
		return
	}
	defer cursor.Close(r.Context())

	var items []models.StockItem
	if err := cursor.All(r.Context(), &items); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode stock items")
		return
	}
	if items == nil {
		items = []models.StockItem{}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"stockItems": items})
}

// CreateStockItem creates a new stock item for a location.
func (h *InventoryHandler) CreateStockItem(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var input struct {
		LocationID   string   `json:"locationId"`
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		Unit         string   `json:"unit"`
		ParLevel     *float64 `json:"parLevel,omitempty"`
		LeadTimeDays int      `json:"leadTimeDays"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	locationID, err := primitive.ObjectIDFromHex(input.LocationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid locationId")
		return
	}

	if input.LeadTimeDays < 1 {
		input.LeadTimeDays = 1
	}

	now := time.Now()
	item := models.StockItem{
		ID:           primitive.NewObjectID(),
		TenantID:     tenantID,
		LocationID:   locationID,
		Name:         input.Name,
		Category:     input.Category,
		Unit:         input.Unit,
		ParLevel:     input.ParLevel,
		LeadTimeDays: input.LeadTimeDays,
		LastModified: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := validation.Validate(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.db.StockItems().InsertOne(r.Context(), item); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			respondWithError(w, http.StatusConflict, "A stock item with this name already exists at this location")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create stock item")
		return
	}

	h.syslog.LogCat(r.Context(), models.LogLow, models.LogCatInventory, "Stock item created: "+item.Name)
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{"stockItem": item})
}

// UpdateStockItem updates an existing stock item.
func (h *InventoryHandler) UpdateStockItem(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	itemIDStr := mux.Vars(r)["id"]
	itemID, err := primitive.ObjectIDFromHex(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid stock item ID")
		return
	}

	var input struct {
		Name         string   `json:"name,omitempty"`
		Category     string   `json:"category,omitempty"`
		Unit         string   `json:"unit,omitempty"`
		ParLevel     *float64 `json:"parLevel,omitempty"`
		LeadTimeDays *int     `json:"leadTimeDays,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	update := bson.M{"lastModified": time.Now(), "updatedAt": time.Now()}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Category != "" {
		update["category"] = input.Category
	}
	if input.Unit != "" {
		update["unit"] = input.Unit
	}
	if input.ParLevel != nil {
		update["parLevel"] = *input.ParLevel
	}
	if input.LeadTimeDays != nil {
		update["leadTimeDays"] = *input.LeadTimeDays
	}

	result, err := h.db.StockItems().UpdateOne(r.Context(),
		bson.M{"_id": itemID, "tenantId": tenantID},
		bson.M{"$set": update})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update stock item")
		return
	}
	if result.MatchedCount == 0 {
		respondWithError(w, http.StatusNotFound, "Stock item not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock item updated"})
}

// DeleteStockItem deletes a stock item.
func (h *InventoryHandler) DeleteStockItem(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	itemIDStr := mux.Vars(r)["id"]
	itemID, err := primitive.ObjectIDFromHex(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid stock item ID")
		return
	}

	result, err := h.db.StockItems().DeleteOne(r.Context(),
		bson.M{"_id": itemID, "tenantId": tenantID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete stock item")
		return
	}
	if result.DeletedCount == 0 {
		respondWithError(w, http.StatusNotFound, "Stock item not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock item deleted"})
}

// SubmitStockCount creates a new stock count for a shift.
func (h *InventoryHandler) SubmitStockCount(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	userID, ok := h.getUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var input struct {
		LocationID string             `json:"locationId"`
		Shift      string             `json:"shift"`
		Counts     []json.RawMessage  `json:"counts"`
		Notes      string             `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	locationID, err := primitive.ObjectIDFromHex(input.LocationID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid locationId")
		return
	}

	if input.Shift == "" {
		input.Shift = "close"
	}
	if len(input.Counts) == 0 {
		respondWithError(w, http.StatusBadRequest, "At least one count is required")
		return
	}

	now := time.Now()
	var entries []models.StockCountEntry
	for _, raw := range input.Counts {
		var entry struct {
			StockItemID string   `json:"stockItemId"`
			Quantity    float64  `json:"quantity"`
			Unit        string   `json:"unit"`
			Received    *float64 `json:"received,omitempty"`
			Waste       *float64 `json:"waste,omitempty"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid count entry format")
			return
		}
		siID, err := primitive.ObjectIDFromHex(entry.StockItemID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid stockItemId in count entry")
			return
		}
		entries = append(entries, models.StockCountEntry{
			StockItemID: siID,
			Quantity:    entry.Quantity,
			Unit:        entry.Unit,
			Received:    entry.Received,
			Waste:       entry.Waste,
		})
	}

	count := models.StockCount{
		ID:          primitive.NewObjectID(),
		TenantID:    tenantID,
		LocationID:  locationID,
		CountedBy:   userID,
		Shift:       input.Shift,
		Counts:      entries,
		Notes:       input.Notes,
		SubmittedAt: now,
		CreatedAt:   now,
	}

	if err := validation.Validate(&count); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.db.StockCounts().InsertOne(r.Context(), count); err != nil {
		// Last-writer-wins: log conflict if duplicate shift detected
		if mongo.IsDuplicateKeyError(err) {
			h.syslog.LogCat(r.Context(), models.LogMedium, models.LogCatInventory,
				"Stock count conflict — duplicate shift submission")
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to submit stock count")
		return
	}

	h.syslog.LogCat(r.Context(), models.LogLow, models.LogCatInventory,
		"Stock count submitted: "+locationID.Hex()+" shift="+input.Shift+" items="+string(len(entries)))
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{"stockCount": count})
}

// GetStockCount retrieves the stock count for a location on a given date.
func (h *InventoryHandler) GetStockCount(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	locationIDStr := r.URL.Query().Get("location_id")
	dateStr := r.URL.Query().Get("date")

	if locationIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	locationID, err := primitive.ObjectIDFromHex(locationIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location_id")
		return
	}

	filter := bson.M{"tenantId": tenantID, "locationId": locationID}
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
		startOfDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)
		filter["submittedAt"] = bson.M{"$gte": startOfDay, "$lt": endOfDay}
	}

	var count models.StockCount
	err = h.db.StockCounts().FindOne(r.Context(), filter, options.FindOne().SetSort(bson.D{{Key: "submittedAt", Value: -1}})).Decode(&count)
	if err == mongo.ErrNoDocuments {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"stockCount": nil})
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch stock count")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"stockCount": count})
}

// GetStockCountHistory returns the last 30 days of stock counts for a specific item.
func (h *InventoryHandler) GetStockCountHistory(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	locationIDStr := r.URL.Query().Get("location_id")
	itemIDStr := r.URL.Query().Get("item_id")
	if itemIDStr == "" || locationIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "item_id and location_id are required")
		return
	}

	itemID, err := primitive.ObjectIDFromHex(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid item_id")
		return
	}
	locationID, err := primitive.ObjectIDFromHex(locationIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location_id")
		return
	}

	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenantId": tenantID, "locationId": locationID, "submittedAt": bson.M{"$gte": thirtyDaysAgo}}}},
		{{Key: "$unwind", Value: "$counts"}},
		{{Key: "$match", Value: bson.M{"counts.stockItemId": itemID}}},
		{{Key: "$sort", Value: bson.M{"submittedAt": -1}}},
		{{Key: "$limit", Value: 30}},
		{{Key: "$project", Value: bson.M{
			"stockItemId": "$counts.stockItemId",
			"quantity":    "$counts.quantity",
			"unit":        "$counts.unit",
			"received":    "$counts.received",
			"waste":       "$counts.waste",
			"submittedAt": 1,
		}}},
	}

	cursor, err := h.db.StockCounts().Aggregate(r.Context(), pipeline)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch history")
		return
	}
	defer cursor.Close(r.Context())

	var history []bson.M
	if err := cursor.All(r.Context(), &history); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode history")
		return
	}
	if history == nil {
		history = []bson.M{}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"history": history})
}

// GetForecast computes consumption forecasts via aggregation pipeline.
func (h *InventoryHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := h.getTenantID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	locationIDStr := r.URL.Query().Get("location_id")
	if locationIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	locationID, err := primitive.ObjectIDFromHex(locationIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid location_id")
		return
	}

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	// Aggregation pipeline: compute 7-day average consumption per item
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"tenantId":    tenantID,
			"locationId":  locationID,
			"submittedAt": bson.M{"$gte": sevenDaysAgo},
		}}},
		{{Key: "$sort", Value: bson.M{"submittedAt": -1}}},
		{{Key: "$unwind", Value: "$counts"}},
		{{Key: "$group", Value: bson.M{
			"_id":       "$counts.stockItemId",
			"avgQty":    bson.M{"$avg": "$counts.quantity"},
			"lastQty":   bson.M{"$first": "$counts.quantity"},
			"lastUnit":  bson.M{"$first": "$counts.unit"},
			"countDays": bson.M{"$sum": 1},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "stock_items",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "item",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$item", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$project", Value: bson.M{
			"stockItemId": "$_id",
			"itemName":    "$item.name",
			"category":    "$item.category",
			"unit":        "$lastUnit",
			"7dayAvg":     bson.M{"$round": bson.A{"$avgQty", 1}},
			"lastQty":     1,
			"countDays":   1,
			"parLevel":    "$item.parLevel",
			"leadTimeDays": bson.M{"$ifNull": bson.A{"$item.leadTimeDays", 1}},
		}}},
	}

	cursor, err := h.db.StockCounts().Aggregate(r.Context(), pipeline)
	if err != nil {
		h.syslog.LogCat(r.Context(), models.LogHigh, models.LogCatInventory,
			"Forecast aggregation failed: "+err.Error())
		respondWithError(w, http.StatusInternalServerError, "Failed to compute forecast")
		return
	}
	defer cursor.Close(r.Context())

	type forecastItem struct {
		StockItemID  primitive.ObjectID `json:"stockItemId" bson:"stockItemId"`
		ItemName     string             `json:"itemName" bson:"itemName"`
		Category     string             `json:"category" bson:"category"`
		Unit         string             `json:"unit" bson:"unit"`
		AvgQty       float64            `json:"7dayAvg" bson:"7dayAvg"`
		LastQty      float64            `json:"lastQty" bson:"lastQty"`
		CountDays    int                `json:"countDays" bson:"countDays"`
		ParLevel     *float64           `json:"parLevel" bson:"parLevel"`
		LeadTimeDays int                `json:"leadTimeDays" bson:"leadTimeDays"`
	}

	var forecast []forecastItem
	if err := cursor.All(r.Context(), &forecast); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode forecast")
		return
	}

	// Compute suggested order quantities
	type forecastResult struct {
		StockItemID       primitive.ObjectID `json:"stockItemId"`
		ItemName          string             `json:"itemName"`
		Category          string             `json:"category"`
		Unit              string             `json:"unit"`
		SevenDayAvg       float64            `json:"7dayAvg"`
		LastQty           float64            `json:"lastQty"`
		CountDays         int                `json:"countDays"`
		SuggestedOrderQty float64            `json:"suggestedOrderQty"`
		ParLevel          *float64           `json:"parLevel"`
	}

	results := make([]forecastResult, 0, len(forecast))
	for _, f := range forecast {
		safetyBuffer := f.AvgQty * 0.2
		suggested := (f.AvgQty * float64(f.LeadTimeDays)) - f.LastQty + safetyBuffer
		if suggested < 0 {
			suggested = 0
		}

		// Cold start: use par_level if not enough data
		if f.CountDays < 2 {
			if f.ParLevel != nil {
				suggested = *f.ParLevel
			} else {
				suggested = 0
			}
		}

		results = append(results, forecastResult{
			StockItemID:       f.StockItemID,
			ItemName:          f.ItemName,
			Category:          f.Category,
			Unit:              f.Unit,
			SevenDayAvg:       f.AvgQty,
			LastQty:           f.LastQty,
			CountDays:         f.CountDays,
			SuggestedOrderQty: suggested,
			ParLevel:          f.ParLevel,
		})
	}

	if results == nil {
		results = []forecastResult{}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"forecast": results})
}
