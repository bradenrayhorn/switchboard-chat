package database

import (
	"fmt"
	"github.com/Kamva/mgm/v3"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Setup() {
	database := viper.GetString("mongo_database")
	err := mgm.SetDefaultConfig(nil, database, options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			viper.GetString("mongo_username"),
			viper.GetString("mongo_password"),
			viper.GetString("mongo_host"),
			viper.GetString("mongo_port"),
			database,
		),
	))

	if err != nil {
		panic(err)
	}

	_, client, _, err := mgm.DefaultConfigs()

	if err != nil {
		panic(err)
	}

	err = client.Ping(mgm.Ctx(), readpref.Primary())

	if err != nil {
		panic(err)
	}

}
