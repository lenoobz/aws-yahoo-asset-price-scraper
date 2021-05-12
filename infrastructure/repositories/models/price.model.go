package models

import (
	"github.com/hthl85/aws-yahoo-price-scraper/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PriceModel struct
type PriceModel struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty"`
	IsActive   bool                `bson:"isActive,omitempty"`
	CreatedAt  int64               `bson:"createdAt,omitempty"`
	ModifiedAt int64               `bson:"modifiedAt,omitempty"`
	Schema     string              `bson:"schema,omitempty"`
	Source     string              `bson:"source,omitempty"`
	Ticker     string              `bson:"ticker,omitempty"`
	Currency   string              `bson:"currency,omitempty"`
	Price      float64             `bson:"price,omitempty"`
}

// NewPriceModel create price model
func NewPriceModel(e *entities.Price) (*PriceModel, error) {
	var m = &PriceModel{}

	m.Ticker = e.Ticker
	m.Currency = e.Currency
	m.Price = e.Price

	return m, nil
}
