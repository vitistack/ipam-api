package mongodbtypes

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	ServiceName         string     `json:"service_name" bson:"service_name" validate:"required" example:"service1"`
	NamespaceId         string     `json:"namespace_id" bson:"namespace_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ClusterId           string     `json:"cluster_id" bson:"cluster_id" validate:"required,min=8,max=64"`
	RetentionPeriodDays int        `json:"retention_period_days" bson:"retention_period_days"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	DenyExternalCleanup bool       `json:"deny_external_cleanup" bson:"deny_external_cleanup"`
}

type Address struct {
	ID       bson.ObjectID `json:"-" bson:"_id"`
	Secret   string        `json:"secret" bson:"secret"`
	Zone     string        `json:"zone" bson:"zone"`
	NetboxID int           `json:"-" bson:"netbox_id"`
	Address  string        `json:"address" bson:"address"`
	Services []Service     `json:"services" bson:"services"`
}
