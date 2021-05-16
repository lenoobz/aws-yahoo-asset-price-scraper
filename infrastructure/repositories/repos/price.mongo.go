package repos

import (
	"context"
	"fmt"
	"time"

	logger "github.com/hthl85/aws-lambda-logger"
	"github.com/hthl85/aws-yahoo-price-scraper/config"
	"github.com/hthl85/aws-yahoo-price-scraper/consts"
	"github.com/hthl85/aws-yahoo-price-scraper/entities"
	"github.com/hthl85/aws-yahoo-price-scraper/infrastructure/repositories/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PriceMongo struct
type PriceMongo struct {
	db     *mongo.Database
	client *mongo.Client
	log    logger.ContextLog
	conf   *config.MongoConfig
}

// NewPriceMongo creates new price mongo repo
func NewPriceMongo(db *mongo.Database, log logger.ContextLog, conf *config.MongoConfig) (*PriceMongo, error) {
	if db != nil {
		return &PriceMongo{
			db:   db,
			log:  log,
			conf: conf,
		}, nil
	}

	// set context with timeout from the config
	// create new context for the query
	ctx, cancel := createContext(context.Background(), conf.TimeoutMS)
	defer cancel()

	// set mongo client options
	clientOptions := options.Client()

	// set min pool size
	if conf.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(conf.MinPoolSize)
	}

	// set max pool size
	if conf.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(conf.MaxPoolSize)
	}

	// set max idle time ms
	if conf.MaxIdleTimeMS > 0 {
		clientOptions.SetMaxConnIdleTime(time.Duration(conf.MaxIdleTimeMS) * time.Millisecond)
	}

	// construct a connection string from mongo config object
	cxnString := fmt.Sprintf("mongodb+srv://%s:%s@%s", conf.Username, conf.Password, conf.Host)

	// create mongo client by making new connection
	client, err := mongo.Connect(ctx, clientOptions.ApplyURI(cxnString))
	if err != nil {
		return nil, err
	}

	return &PriceMongo{
		db:     client.Database(conf.Dbname),
		client: client,
		log:    log,
		conf:   conf,
	}, nil
}

// Close disconnect from database
func (r *PriceMongo) Close() {
	ctx := context.Background()
	r.log.Info(ctx, "close mongo client")

	if r.client == nil {
		return
	}

	if err := r.client.Disconnect(ctx); err != nil {
		r.log.Error(ctx, "disconnect mongo failed", "error", err)
	}
}

///////////////////////////////////////////////////////////////////////////////
// Implement interface
///////////////////////////////////////////////////////////////////////////////

// CountAssets count number of assets available
func (r *PriceMongo) CountAssets(ctx context.Context) (int64, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.ASSETS_COLLECTION]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return 0, fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	// filter
	filter := bson.D{}

	// find options
	countOptions := options.Count()

	return col.CountDocuments(ctx, filter, countOptions)
}

// FindAllAssets find all assets
func (r *PriceMongo) FindAllAssets(ctx context.Context) ([]*entities.Asset, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.ASSETS_COLLECTION]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return nil, fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	// filter
	filter := bson.D{}

	// find options
	findOptions := options.Find()

	cur, err := col.Find(ctx, filter, findOptions)

	// only run defer function when find success
	if cur != nil {
		defer func() {
			if deferErr := cur.Close(ctx); deferErr != nil {
				err = deferErr
			}
		}()
	}

	// find was not succeed
	if err != nil {
		r.log.Error(ctx, "find query failed", "error", err)
		return nil, err
	}

	var assets []*entities.Asset

	// iterate over the cursor to decode document one at a time
	for cur.Next(ctx) {
		// decode cursor to activity model
		var asset entities.Asset
		if err = cur.Decode(&asset); err != nil {
			r.log.Error(ctx, "decode failed", "error", err)
			return nil, err
		}

		assets = append(assets, &asset)
	}

	if err := cur.Err(); err != nil {
		r.log.Error(ctx, "iterate over cursor failed", "error", err)
		return nil, err
	}

	return assets, nil
}

// FindAssetsFromCheckpoint find assets from checkpoint
func (r *PriceMongo) FindAssetsFromCheckpoint(ctx context.Context, checkpoint *entities.Checkpoint) ([]*entities.Asset, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.ASSETS_COLLECTION]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return nil, fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	// filter
	filter := bson.D{}

	// find options
	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(checkpoint.PageIndex * checkpoint.PageSize).SetLimit(checkpoint.PageSize)

	cur, err := col.Find(ctx, filter, findOptions)

	// only run defer function when find success
	if cur != nil {
		defer func() {
			if deferErr := cur.Close(ctx); deferErr != nil {
				err = deferErr
			}
		}()
	}

	// find was not succeed
	if err != nil {
		r.log.Error(ctx, "find query failed", "error", err)
		return nil, err
	}

	var assets []*entities.Asset

	// iterate over the cursor to decode document one at a time
	for cur.Next(ctx) {
		// decode cursor to activity model
		var asset entities.Asset
		if err = cur.Decode(&asset); err != nil {
			r.log.Error(ctx, "decode failed", "error", err)
			return nil, err
		}

		assets = append(assets, &asset)
	}

	if err := cur.Err(); err != nil {
		r.log.Error(ctx, "iterate over cursor failed", "error", err)
		return nil, err
	}

	return assets, nil
}

// UpdateCheckpoint updates a checkpoint given page size and number of asset
func (r *PriceMongo) UpdateCheckpoint(ctx context.Context, pageSize int64, numAssets int64) (*entities.Checkpoint, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.SCRAPE_CHECKPOINT_COLLECTION]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return nil, fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	// filter
	filter := bson.D{}

	// find options
	findOptions := options.FindOne()

	// decode cursor to activity model
	var checkpoint models.CheckPointModel
	checkpoint.PageSize = pageSize

	cur := col.FindOne(ctx, filter, findOptions)

	// only run defer function when find success
	err := cur.Err()

	if err == mongo.ErrNoDocuments {
		checkpoint.PrevIndex = 0
		return r.updateCheckPoint(ctx, col, &checkpoint)
	}

	// find was not succeed
	if err != nil {
		r.log.Error(ctx, "find query failed", "error", err)
		return nil, err
	}

	if err = cur.Decode(&checkpoint); err != nil {
		r.log.Error(ctx, "decode failed", "error", err)
		return nil, err
	}

	if checkpoint.PrevIndex*checkpoint.PageSize+checkpoint.PageSize >= numAssets {
		checkpoint.PrevIndex = 0
	} else {
		checkpoint.PrevIndex = checkpoint.PrevIndex + 1
	}

	return r.updateCheckPoint(ctx, col, &checkpoint)
}

// InsertAssetPrice insert asset price
func (r *PriceMongo) InsertAssetPrice(ctx context.Context, assetPrice *entities.Price) error {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	priceModel, err := models.NewPriceModel(assetPrice)
	if err != nil {
		r.log.Error(ctx, "create model failed", "error", err)
		return err
	}

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.ASSET_PRICES_COLLECTION]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	priceModel.IsActive = true
	priceModel.Schema = r.conf.SchemaVersion
	priceModel.ModifiedAt = time.Now().UTC().Unix()

	filter := bson.D{{
		Key:   "ticker",
		Value: priceModel.Ticker,
	}}

	update := bson.D{
		{
			Key:   "$set",
			Value: priceModel,
		},
		{
			Key: "$setOnInsert",
			Value: bson.D{{
				Key:   "createdAt",
				Value: time.Now().UTC().Unix(),
			}},
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err = col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.log.Error(ctx, "update one failed", "error", err)
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////
// Implement helper function
///////////////////////////////////////////////////////////

// createContext create a new context with timeout
func createContext(ctx context.Context, timeout uint64) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
}

// updateCheckPoint update checkpoint
func (r *PriceMongo) updateCheckPoint(ctx context.Context, col *mongo.Collection, checkpoint *models.CheckPointModel) (*entities.Checkpoint, error) {
	checkpoint.IsActive = true
	checkpoint.Schema = r.conf.SchemaVersion
	checkpoint.ModifiedAt = time.Now().UTC().Unix()

	// filter
	filter := bson.D{}
	if checkpoint.ID != nil {
		filter = bson.D{{Key: "_id", Value: checkpoint.ID}}
	} else {
		checkpoint.CreatedAt = time.Now().UTC().Unix()
	}

	// update
	update := bson.D{
		{
			Key:   "$set",
			Value: checkpoint,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.log.Error(ctx, "update one failed", "error", err)
		return nil, err
	}

	return &entities.Checkpoint{
		PageSize:  checkpoint.PageSize,
		PageIndex: checkpoint.PrevIndex,
	}, nil
}
