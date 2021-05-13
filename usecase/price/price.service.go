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

// GetAllAssets gets all assets
func (s *Service) GetAllAssets(ctx context.Context) ([]*entities.Asset, error) {
	s.log.Info(ctx, "get all assets")
	return s.repo.FindAllAssets(ctx)
}

// GetAssetsFromCheckpoint gets all assets from checkpoint
func (s *Service) GetAssetsFromCheckpoint(ctx context.Context, pageSize int64) ([]*entities.Asset, error) {
	s.log.Info(ctx, "get assets from checkpoint")
	numAssets, err := s.repo.CountAssets(ctx)
	if err != nil {
		s.log.Error(ctx, "count assets failed", "error", err)
	}

	checkpoint, err := s.repo.UpdateCheckpoint(ctx, pageSize, numAssets)
	if err != nil {
		s.log.Error(ctx, "find and update checkpoint failed", "error", err)
	}

	if checkpoint == nil {
		s.log.Error(ctx, "checkpoint is nil", "checkpoint", checkpoint)
		return nil, nil
	}

	return s.repo.FindAssetsFromCheckpoint(ctx, checkpoint)
}

// AddAssetPrice creates new asset price
func (s *Service) AddAssetPrice(ctx context.Context, assetPrice *entities.Price) error {
	s.log.Info(ctx, "adding asset price", "ticker", assetPrice.Ticker)
	return s.repo.InsertAssetPrice(ctx, assetPrice)
}
