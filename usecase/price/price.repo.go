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
	CountSecurites(context.Context) (int64, error)
	FindSecurities(context.Context) ([]*entities.Security, error)
	FindSecuritiesPaging(context.Context, *entities.CheckPoint) ([]*entities.Security, error)
}

// Writer interface
type Writer interface {
	InsertPrice(context.Context, *entities.Price) error
	FindOneAndUpdateCheckPoint(ctx context.Context, pageSize int64, numSecurities int64) (*entities.CheckPoint, error)
}

// Repo interface
type Repo interface {
	Reader
	Writer
}
