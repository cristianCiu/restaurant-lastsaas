package validation

import (
	"testing"
	"time"

	"lastsaas/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validStockItem() models.StockItem {
	return models.StockItem{
		TenantID:     primitive.NewObjectID(),
		LocationID:   primitive.NewObjectID(),
		Name:         "Buns",
		Category:     "Bakery",
		Unit:         "cases",
		LeadTimeDays: 1,
		LastModified: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestValidate_ValidStockItem(t *testing.T) {
	item := validStockItem()
	if err := Validate(&item); err != nil {
		t.Errorf("expected valid stock item to pass: %v", err)
	}
}

func TestValidate_StockItemMissingName(t *testing.T) {
	item := validStockItem()
	item.Name = ""
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestValidate_StockItemMissingCategory(t *testing.T) {
	item := validStockItem()
	item.Category = ""
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for missing category")
	}
}

func TestValidate_StockItemMissingUnit(t *testing.T) {
	item := validStockItem()
	item.Unit = ""
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for missing unit")
	}
}

func TestValidate_StockItemInvalidLeadTime(t *testing.T) {
	item := validStockItem()
	item.LeadTimeDays = 0
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for lead time < 1")
	}
}

func TestValidate_StockItemMissingTenant(t *testing.T) {
	item := validStockItem()
	item.TenantID = primitive.NilObjectID
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for missing tenantId")
	}
}

func TestValidate_StockItemNameTooLong(t *testing.T) {
	item := validStockItem()
	item.Name = ""
	for i := 0; i < 210; i++ {
		item.Name += "a"
	}
	if err := Validate(&item); err == nil {
		t.Fatal("expected validation error for name > 200 chars")
	}
}

func validStockCount() models.StockCount {
	return models.StockCount{
		TenantID:   primitive.NewObjectID(),
		LocationID: primitive.NewObjectID(),
		CountedBy:  primitive.NewObjectID(),
		Shift:      "close",
		Counts: []models.StockCountEntry{
			{
				StockItemID: primitive.NewObjectID(),
				Quantity:    10,
				Unit:        "cases",
			},
		},
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
	}
}

func TestValidate_ValidStockCount(t *testing.T) {
	count := validStockCount()
	if err := Validate(&count); err != nil {
		t.Errorf("expected valid stock count to pass: %v", err)
	}
}

func TestValidate_StockCountMissingShift(t *testing.T) {
	count := validStockCount()
	count.Shift = ""
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for missing shift")
	}
}

func TestValidate_StockCountInvalidShift(t *testing.T) {
	count := validStockCount()
	count.Shift = "morning"
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for invalid shift")
	}
}

func TestValidate_StockCountEmptyCounts(t *testing.T) {
	count := validStockCount()
	count.Counts = []models.StockCountEntry{}
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for empty counts")
	}
}

func TestValidate_StockCountMissingCountedBy(t *testing.T) {
	count := validStockCount()
	count.CountedBy = primitive.NilObjectID
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for missing countedBy")
	}
}

func TestValidate_StockCountEntryNegativeQuantity(t *testing.T) {
	count := validStockCount()
	count.Counts[0].Quantity = -1
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for negative quantity")
	}
}

func TestValidate_StockCountEntryMissingUnit(t *testing.T) {
	count := validStockCount()
	count.Counts[0].Unit = ""
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for missing unit in entry")
	}
}

func TestValidate_StockCountTooManyEntries(t *testing.T) {
	count := validStockCount()
	entries := make([]models.StockCountEntry, 501)
	for i := 0; i < 501; i++ {
		entries[i] = models.StockCountEntry{
			StockItemID: primitive.NewObjectID(),
			Quantity:    1,
			Unit:        "cases",
		}
	}
	count.Counts = entries
	if err := Validate(&count); err == nil {
		t.Fatal("expected validation error for > 500 entries")
	}
}
