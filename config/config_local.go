// +build local

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
		Dbname:        "etf_funds_local",
		SchemaVersion: "1",
		Colnames: map[string]string{
			"stock":      "securities",
			"price":      "prices",
			"checkpoint": "checkpoint",
		},
	},
}
