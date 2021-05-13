package price

import (
	"context"

	"github.com/hthl85/aws-yahoo-price-scraper/entities"
)

///////////////////////////////////////////////////////////
// Price Repository Interface
///////////////////////////////////////////////////////////

// Reader interface
type Reader interface {
	CountAssets(ctx context.Context) (int64, error)
	FindAllAssets(ctx context.Context) ([]*entities.Asset, error)
	FindAssetsFromCheckpoint(ctx context.Context, checkpoint *entities.Checkpoint) ([]*entities.Asset, error)
}

// Writer interface
type Writer interface {
	InsertAssetPrice(ctx context.Context, assetPrice *entities.Price) error
	UpdateCheckpoint(ctx context.Context, pageSize int64, numAssets int64) (*entities.Checkpoint, error)
}

// Repo interface
type Repo interface {
	Reader
	Writer
}
