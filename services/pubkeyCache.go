package services

import (
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

var pubkeyCacheDb *leveldb.DB

func initPubkeyCache(path string) error {
	if path == "" {
		logger.Infof("no last pubkey cache path provided, using temporary directory %v", os.TempDir()+"/pubkeyCache")
		path = os.TempDir() + "/pubkeyCache"
	}
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return err
	}

	pubkeyCacheDb = db
	return nil
}
