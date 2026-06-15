package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockItem struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TenantID     primitive.ObjectID `json:"tenantId" bson:"tenantId" validate:"required"`
	LocationID   primitive.ObjectID `json:"locationId" bson:"locationId" validate:"required"`
	Name         string             `json:"name" bson:"name" validate:"required,min=1,max=200"`
	Category     string             `json:"category" bson:"category" validate:"required,min=1,max=100"`
	Unit         string             `json:"unit" bson:"unit" validate:"required,min=1,max=50"`
	ParLevel     *float64           `json:"parLevel,omitempty" bson:"parLevel,omitempty"`
	LeadTimeDays int                `json:"leadTimeDays" bson:"leadTimeDays" validate:"gte=1"`
	LastModified time.Time          `json:"lastModified" bson:"lastModified" validate:"required"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt" validate:"required"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt" validate:"required"`
}
