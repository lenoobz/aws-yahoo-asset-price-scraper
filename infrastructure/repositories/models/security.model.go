package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// SecurityModel struct
type SecurityModel struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty"`
	IsActive   bool                `bson:"isActive,omitempty"`
	CreatedAt  int64               `bson:"createdAt,omitempty"`
	ModifiedAt int64               `bson:"modifiedAt,omitempty"`
	Schema     string              `bson:"schema,omitempty"`
	Source     string              `bson:"source,omitempty"`
	Ticker     string              `bson:"ticker,omitempty"`
	Currency   string              `bson:"currency,omitempty"`
}
