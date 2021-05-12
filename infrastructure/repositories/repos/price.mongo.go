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
func NewPriceMongo(db *mongo.Database, l logger.ContextLog, conf *config.MongoConfig) (*PriceMongo, error) {
	if db != nil {
		return &PriceMongo{
			db:   db,
			log:  l,
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
		log:    l,
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

// FindOneAndUpdateCheckPoint find a check point and update it value
func (r *PriceMongo) FindOneAndUpdateCheckPoint(ctx context.Context, pageSize int64, numSecurities int64) (*entities.CheckPoint, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.CHECK_POINT_COL]
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
	var cp models.CheckPointModel
	cp.PageSize = pageSize

	cur := col.FindOne(ctx, filter, findOptions)

	// only run defer function when find success
	err := cur.Err()

	if err == mongo.ErrNoDocuments {
		cp.PrevIndex = 0
		return r.updateCheckPoint(ctx, col, &cp)
	}

	// find was not succeed
	if err != nil {
		r.log.Error(ctx, "find query failed", "error", err)
		return nil, err
	}

	if err = cur.Decode(&cp); err != nil {
		r.log.Error(ctx, "decode failed", "error", err)
		return nil, err
	}

	if cp.PrevIndex*cp.PageSize+cp.PageSize >= numSecurities {
		cp.PrevIndex = 0
	} else {
		cp.PrevIndex = cp.PrevIndex + 1
	}

	return r.updateCheckPoint(ctx, col, &cp)
}

func (r *PriceMongo) updateCheckPoint(ctx context.Context, col *mongo.Collection, m *models.CheckPointModel) (*entities.CheckPoint, error) {

	m.IsActive = true
	m.Schema = r.conf.SchemaVersion
	m.ModifiedAt = time.Now().UTC().Unix()

	// filter
	filter := bson.D{}
	if m.ID != nil {
		filter = bson.D{{Key: "_id", Value: m.ID}}
	} else {
		m.CreatedAt = time.Now().UTC().Unix()
	}

	// update
	update := bson.D{
		{
			Key:   "$set",
			Value: m,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.log.Error(ctx, "update one failed", "error", err)
		return nil, err
	}

	cp := entities.CheckPoint{
		PageSize:  m.PageSize,
		PageIndex: m.PrevIndex,
	}

	return &cp, nil
}

// CountSecurites count number of securities available
func (r *PriceMongo) CountSecurites(ctx context.Context) (int64, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.SECURITIES_COL]
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

// FindSecurities gets all securities
func (r *PriceMongo) FindSecurities(ctx context.Context) ([]*entities.Security, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.SECURITIES_COL]
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

	var securities []*entities.Security

	// iterate over the cursor to decode document one at a time
	for cur.Next(ctx) {
		// decode cursor to activity model
		var security entities.Security
		if err = cur.Decode(&security); err != nil {
			r.log.Error(ctx, "decode failed", "error", err)
			return nil, err
		}

		securities = append(securities, &security)
	}

	if err := cur.Err(); err != nil {
		r.log.Error(ctx, "iterate over cursor failed", "error", err)
		return nil, err
	}

	return securities, nil
}

// FindSecuritiesPaging gets all securities by paging
func (r *PriceMongo) FindSecuritiesPaging(ctx context.Context, e *entities.CheckPoint) ([]*entities.Security, error) {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.SECURITIES_COL]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return nil, fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	// filter
	filter := bson.D{}

	// find options
	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetSkip(e.PageIndex * e.PageSize).SetLimit(e.PageSize)

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

	var securities []*entities.Security

	// iterate over the cursor to decode document one at a time
	for cur.Next(ctx) {
		// decode cursor to activity model
		var security entities.Security
		if err = cur.Decode(&security); err != nil {
			r.log.Error(ctx, "decode failed", "error", err)
			return nil, err
		}

		securities = append(securities, &security)
	}

	if err := cur.Err(); err != nil {
		r.log.Error(ctx, "iterate over cursor failed", "error", err)
		return nil, err
	}

	return securities, nil
}

// InsertPrice insert price
func (r *PriceMongo) InsertPrice(ctx context.Context, e *entities.Price) error {
	// create new context for the query
	ctx, cancel := createContext(ctx, r.conf.TimeoutMS)
	defer cancel()

	m, err := models.NewPriceModel(e)
	if err != nil {
		r.log.Error(ctx, "create model failed", "error", err)
		return err
	}

	// what collection we are going to use
	colname, ok := r.conf.Colnames[consts.PRICE_COL]
	if !ok {
		r.log.Error(ctx, "cannot find collection name")
		return fmt.Errorf("cannot find collection name")
	}
	col := r.db.Collection(colname)

	m.IsActive = true
	m.Schema = r.conf.SchemaVersion
	m.ModifiedAt = time.Now().UTC().Unix()

	filter := bson.D{{
		Key:   "ticker",
		Value: m.Ticker,
	}}

	update := bson.D{
		{
			Key:   "$set",
			Value: m,
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
func createContext(ctx context.Context, t uint64) (context.Context, context.CancelFunc) {
	timeout := time.Duration(t) * time.Millisecond
	return context.WithTimeout(ctx, timeout*time.Millisecond)
}
