package mongodbtypes

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	ServiceName         string     `json:"service_name" bson:"service_name" validate:"required" example:"service1"`
	ServiceId           string     `json:"service_id" bson:"service_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ClusterId           string     `json:"cluster_id" bson:"cluster_id" validate:"required,min=8,max=64"`
	RetentionPeriodDays int        `json:"retention_period_days" bson:"retention_period_days"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

type Address struct {
	ID       bson.ObjectID `json:"-" bson:"_id"`
	Secret   string        `json:"secret" bson:"secret"`
	Zone     string        `json:"zone" bson:"zone"`
	NetboxID int           `json:"-" bson:"id"`
	Address  string        `json:"address" bson:"address"`
	Services []Service     `json:"services" bson:"services"`
}
