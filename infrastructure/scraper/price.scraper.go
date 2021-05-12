package scraper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/google/uuid"
	corid "github.com/hthl85/aws-lambda-corid"
	logger "github.com/hthl85/aws-lambda-logger"
	"github.com/hthl85/aws-yahoo-price-scraper/config"
	"github.com/hthl85/aws-yahoo-price-scraper/entities"
	"github.com/hthl85/aws-yahoo-price-scraper/usecase/price"
)

// PriceScraper struct
type PriceScraper struct {
	StockJob     *colly.Collector
	priceService *price.Service
	errorTickers []string
	log          logger.ContextLog
}

// NewPriceScraper create new price scraper
func NewPriceScraper(ps *price.Service, l logger.ContextLog) *PriceScraper {
	sj := newScraperJob()

	return &PriceScraper{
		StockJob:     sj,
		priceService: ps,
		log:          l,
	}
}

// newScraperJob creates a new colly collector with some custom configs
func newScraperJob() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains(config.AllowDomain),
		colly.Async(true),
	)

	// Overrides the default timeout (10 seconds) for this collector
	c.SetRequestTimeout(30 * time.Second)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		DomainGlob:  config.DomainGlob,
		Parallelism: 2,
		RandomDelay: 2 * time.Second,
	})

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	return c
}

// configJobs configs on error handler and on response handler for scaper jobs
func (s *PriceScraper) configJobs() {
	s.StockJob.OnError(s.errorHandler)
	s.StockJob.OnScraped(s.scrapedHandler)
	s.StockJob.OnHTML("div[id=quote-header-info]", s.processPriceResponse)
}

// StartJobs start job
func (s *PriceScraper) StartJobs() {
	ctx := context.Background()

	s.configJobs()

	securities, err := s.priceService.GetSecurities(ctx)
	if err != nil {
		s.log.Error(ctx, "get securities list failed", "error", err)
	}

	for _, security := range securities {
		reqContext := colly.NewContext()
		reqContext.Put("ticker", security.Ticker)
		reqContext.Put("currency", security.Currency)

		url := config.GetPriceByTickerURL(security.Ticker)

		s.log.Info(ctx, "scraping price security", "ticker", security.Ticker)

		if err := s.StockJob.Request("GET", url, nil, reqContext, nil); err != nil {
			s.log.Error(ctx, "scrape price security failed", "error", err, "ticker", security.Ticker)
		}
	}

	s.StockJob.Wait()
}

// StartJob start job
func (s *PriceScraper) StartJobsByPaging(pageSize int64) {
	ctx := context.Background()

	s.configJobs()

	securities, err := s.priceService.GetSecuritiesPaging(ctx, pageSize)
	if err != nil {
		s.log.Error(ctx, "get securities list failed", "error", err)
	}

	for _, security := range securities {
		reqContext := colly.NewContext()
		reqContext.Put("ticker", security.Ticker)
		reqContext.Put("currency", security.Currency)

		url := config.GetPriceByTickerURL(security.Ticker)

		s.log.Info(ctx, "scraping price security", "ticker", security.Ticker)

		if err := s.StockJob.Request("GET", url, nil, reqContext, nil); err != nil {
			s.log.Error(ctx, "scrape price security failed", "error", err, "ticker", security.Ticker)
		}
	}

	s.StockJob.Wait()
}

///////////////////////////////////////////////////////////
// Scraper Handler
///////////////////////////////////////////////////////////

// errorHandler generic error handler for all scaper jobs
func (s *PriceScraper) errorHandler(r *colly.Response, err error) {
	ctx := context.Background()
	s.log.Error(ctx, "failed to request url", "url", r.Request.URL, "error", err)
	s.errorTickers = append(s.errorTickers, r.Request.Ctx.Get("ticker"))
}

func (s *PriceScraper) scrapedHandler(r *colly.Response) {
	ctx := context.Background()
	foundPrice := r.Ctx.Get("foundPrice")
	if foundPrice == "" {
		s.log.Error(ctx, "price not found", "ticker", r.Request.Ctx.Get("ticker"))
		s.errorTickers = append(s.errorTickers, r.Request.Ctx.Get("ticker"))
		return
	}
}

func (s *PriceScraper) processPriceResponse(e *colly.HTMLElement) {
	// create correlation if for processing fund list
	id, _ := uuid.NewRandom()
	ctx := corid.NewContext(context.Background(), id)

	ticker := e.Request.Ctx.Get("ticker")
	currency := e.Request.Ctx.Get("currency")
	s.log.Info(ctx, "processPriceResponse", "ticker", ticker)

	foundPrice := false

	stock := entities.Price{
		Ticker:   ticker,
		Currency: currency,
	}

	e.ForEach("span", func(_ int, span *colly.HTMLElement) {
		txt := span.Attr("data-reactid")
		if strings.EqualFold(txt, "32") {
			p := span.DOM.Text()
			val, err := strconv.ParseFloat(p, 64)
			if err != nil {
				s.log.Error(ctx, "parse price failed", "error", err, "ticker", ticker, "raw-value", txt)
				return
			}

			stock.Price = val
			foundPrice = true
		}
	})

	if foundPrice {
		e.Response.Ctx.Put("foundPrice", "true")

		if err := s.priceService.AddPrice(ctx, &stock); err != nil {
			s.log.Error(ctx, "add price failed", "error", err, "ticker", ticker)
		}
	}
}

// Close scraper
func (s *PriceScraper) Close() {
	s.log.Info(context.Background(), "DONE - SCRAPING STOCKS PRICE", "errorTickers", s.errorTickers)
}
