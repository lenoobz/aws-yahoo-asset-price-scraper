package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// CheckPointModel struct
type CheckPointModel struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty"`
	IsActive   bool                `bson:"isActive,omitempty"`
	CreatedAt  int64               `bson:"createdAt,omitempty"`
	ModifiedAt int64               `bson:"modifiedAt,omitempty"`
	Schema     string              `bson:"schema,omitempty"`
	PageSize   int64               `bson:"size,omitempty"`
	PrevIndex  int64               `bson:"prevIndex"`
}
