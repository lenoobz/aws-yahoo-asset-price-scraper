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
func NewPriceModel(assetPrice *entities.Price) (*PriceModel, error) {
	return &PriceModel{
		Ticker:   assetPrice.Ticker,
		Currency: assetPrice.Currency,
		Price:    assetPrice.Price,
	}, nil
}
