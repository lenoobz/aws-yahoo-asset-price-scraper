package price

import (
	"context"

	logger "github.com/hthl85/aws-lambda-logger"
	"github.com/hthl85/aws-yahoo-price-scraper/entities"
)

// Service sector
type Service struct {
	repo Repo
	log  logger.ContextLog
}

// NewService create new service
func NewService(r Repo, l logger.ContextLog) *Service {
	return &Service{
		repo: r,
		log:  l,
	}
}

// GetSecurities gets all securities
func (s *Service) GetSecurities(ctx context.Context) ([]*entities.Security, error) {
	s.log.Info(ctx, "get securities")
	return s.repo.FindSecurities(ctx)
}

// GetSecuritiesPaging gets all securities by paging
func (s *Service) GetSecuritiesPaging(ctx context.Context, pageSize int64) ([]*entities.Security, error) {
	s.log.Info(ctx, "get securities by tickers")
	numSecurities, err := s.repo.CountSecurites(ctx)
	if err != nil {
		s.log.Error(ctx, "count sercurities failed", "error", err)
	}

	cp, err := s.repo.FindOneAndUpdateCheckPoint(ctx, pageSize, numSecurities)
	if err != nil {
		s.log.Error(ctx, "find and update checkpoint failed", "error", err)
	}

	if cp == nil {
		s.log.Error(ctx, "checkpoint is nil", "checkpoint", cp)
		return nil, nil
	}

	return s.repo.FindSecuritiesPaging(ctx, cp)
}

// AddPrice creates new price
func (s *Service) AddPrice(ctx context.Context, e *entities.Price) error {
	s.log.Info(ctx, "adding price", "ticker", e.Ticker)
	return s.repo.InsertPrice(ctx, e)
}
