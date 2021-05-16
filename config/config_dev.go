// +build dev

package config

// AppConf constants
var AppConf = AppConfig{
	Mongo: MongoConfig{
		TimeoutMS:     360000,
		MinPoolSize:   5,
		MaxPoolSize:   10,
		MaxIdleTimeMS: 360000,
		Host:          "lenoobetfdevcluster.jd7wd.mongodb.net",
		Username:      "lenoob_dev",
		Password:      "lenoob_dev",
		Dbname:        "povi_dev",
		SchemaVersion: "1",
		Colnames: map[string]string{
			"assets":            "assets",
			"asset_prices":      "asset_prices",
			"scrape_checkpoint": "scrape_checkpoint",
		},
	},
}
