package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockCountEntry struct {
	StockItemID primitive.ObjectID `json:"stockItemId" bson:"stockItemId" validate:"required"`
	Quantity    float64            `json:"quantity" bson:"quantity" validate:"gte=0"`
	Unit        string             `json:"unit" bson:"unit" validate:"required,min=1,max=50"`
	Received    *float64           `json:"received,omitempty" bson:"received,omitempty"`
	Waste       *float64           `json:"waste,omitempty" bson:"waste,omitempty"`
}

type StockCount struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID     primitive.ObjectID `json:"tenantId" bson:"tenantId" validate:"required"`
	LocationID   primitive.ObjectID `json:"locationId" bson:"locationId" validate:"required"`
	CountedBy    primitive.ObjectID `json:"countedBy" bson:"countedBy" validate:"required"`
	Shift        string             `json:"shift" bson:"shift" validate:"required,oneof=close lunch"`
	Counts       []StockCountEntry  `json:"counts" bson:"counts" validate:"required,min=1,max=500"`
	Notes        string             `json:"notes,omitempty" bson:"notes,omitempty"`
	SubmittedAt  time.Time          `json:"submittedAt" bson:"submittedAt" validate:"required"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt" validate:"required"`
}
