// +build dev

package config

import "os"

var host = os.Getenv("MONGO_DB_HOST")
var username = os.Getenv("MONGO_DB_PASSWORD")
var password = os.Getenv("MONGO_DB_USERNAME")

// AppConf constants
var AppConf = AppConfig{
	Mongo: MongoConfig{
		TimeoutMS:     360000,
		MinPoolSize:   5,
		MaxPoolSize:   10,
		MaxIdleTimeMS: 360000,
		Host:          host,
		Username:      username,
		Password:      password,
		Dbname:        "povi_dev",
		SchemaVersion: "1",
		Colnames: map[string]string{
			"assets":            "assets",
			"asset_prices":      "asset_prices",
			"scrape_checkpoint": "scrape_checkpoint",
		},
	},
}
