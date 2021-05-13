package main

import (
	"log"

	logger "github.com/hthl85/aws-lambda-logger"
	"github.com/hthl85/aws-yahoo-price-scraper/config"
	"github.com/hthl85/aws-yahoo-price-scraper/consts"
	"github.com/hthl85/aws-yahoo-price-scraper/infrastructure/repositories/repos"
	"github.com/hthl85/aws-yahoo-price-scraper/infrastructure/scraper"
	"github.com/hthl85/aws-yahoo-price-scraper/usecase/price"
)

func main() {
	appConf := config.AppConf

	// create new logger
	zap, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal("create app logger failed")
	}
	defer zap.Close()

	// create new repository
	repo, err := repos.NewPriceMongo(nil, zap, &appConf.Mongo)
	if err != nil {
		log.Fatal("create price mongo repo failed")
	}
	defer repo.Close()

	// create new services
	ps := price.NewService(repo, zap)

	ts := scraper.NewPriceScraper(ps, zap)
	// ts.ScrapeAllAssetsPrice()
	ts.ScrapeAssetsPriceFromCheckpoint(consts.PAGE_SIZE)
	defer ts.Close()
}
