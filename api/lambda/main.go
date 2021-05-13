package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
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
		log.Fatal("create stock mongo repo failed")
	}
	defer repo.Close()

	// create new service
	fs := price.NewService(repo, zap)

	// create new scraper jobs
	jobs := scraper.NewPriceScraper(fs, zap)
	jobs.ScrapeAssetsPriceFromCheckpoint(consts.PAGE_SIZE)
	defer jobs.Close()

	lambda.Start(lambdaHandler)
}

func lambdaHandler() {
	log.Println("lambda handler is called")
}
