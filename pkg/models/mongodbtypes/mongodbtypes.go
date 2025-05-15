package mongodbtypes

import "go.mongodb.org/mongo-driver/v2/bson"

type Service struct {
	Name     string `json:"name"`
	Uuid     string `json:"uuid"`
	Location string `json:"location"`
}

type Prefix struct {
	ID       bson.ObjectID `json:"-" bson:"_id"`
	Secret   string        `json:"secret" bson:"secret"`
	Zone     string        `json:"zone" bson:"zone"`
	NetboxID int           `json:"-" bson:"id"`
	Prefix   string        `json:"prefix" bson:"prefix"`
	Services []Service     `json:"services" bson:"services"`
}
