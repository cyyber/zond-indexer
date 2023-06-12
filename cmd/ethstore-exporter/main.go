package main

import (
	"flag"

	"github.com/Prajjawalk/zond-indexer/db"
	"github.com/Prajjawalk/zond-indexer/exporter"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/Prajjawalk/zond-indexer/version"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	bnAddress := flag.String("beacon-node-address", "https://falling-powerful-knowledge.ethereum-goerli.discover.quiknode.pro/44a07fc5f81690972c4e4ad4e6bc466f6a1c9b98", "Url of the beacon node api")
	enAddress := flag.String("execution-node-address", "https://goerli.infura.io/v3/d1b7a32b15534fe593f207a0981f930b", "Url of the execution node api")
	updateInterval := flag.Duration("update-intv", 0, "Update interval")
	errorInterval := flag.Duration("error-intv", 0, "Error interval")
	sleepInterval := flag.Duration("sleep-intv", 0, "Sleep interval")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	db.MustInitDB(&types.DatabaseConfig{
		Username: cfg.WriterDatabase.Username,
		Password: cfg.WriterDatabase.Password,
		Name:     cfg.WriterDatabase.Name,
		Host:     cfg.WriterDatabase.Host,
		Port:     cfg.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.ReaderDatabase.Username,
		Password: cfg.ReaderDatabase.Password,
		Name:     cfg.ReaderDatabase.Name,
		Host:     cfg.ReaderDatabase.Host,
		Port:     cfg.ReaderDatabase.Port,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	exporter.StartEthStoreExporter(*bnAddress, *enAddress, *updateInterval, *errorInterval, *sleepInterval)
	logrus.Println("exiting...")
}
