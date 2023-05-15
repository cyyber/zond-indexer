package main

import (
	"context"
	"flag"
	"log"
	"strconv"

	"github.com/Prajjawalk/zond-indexer/db"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	mongo, err := db.InitMongodb(utils.Config.MongoDB.ConnectionString, utils.Config.MongoDB.Instance, strconv.FormatUint(utils.Config.Chain.Config.DepositChainID, 10))
	if err != nil {
		log.Fatal(err)
	}

	collectionList := []string{"beaconchain", "blocks", "cache", "data", "machine_metrics", "metadata", "metadata_updates"}

	for _, i := range collectionList {
		err = mongo.Db.CreateCollection(context.Background(), i)
		if err != nil {
			log.Fatal(err)
		}
	}

	defer mongo.Close()

}
