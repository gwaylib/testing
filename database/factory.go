package database

import (
	"sync"

	"github.com/gwaylib/datastore/conf/etc"
	"github.com/gwaylib/errors"
)

var (
	cacheLock = sync.Mutex{}
	cache     = map[string]*DB{}
)

func getDB(etcFileName, sectionName string) (*DB, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	key := etcFileName + sectionName

	// get from cache
	db, ok := cache[key]
	if ok {
		return db, nil
	}

	// create a new
	cfg, err := etc.GetEtc(etcFileName)
	if err != nil {
		return nil, errors.As(err, etcFileName)
	}
	drvName := cfg.Section(sectionName).Key("driver").String()
	dsn := cfg.Section(sectionName).Key("dsn").String()
	db, err = Open(drvName, dsn)
	if err != nil {
		return nil, errors.As(err)
	}
	cache[key] = db
	return db, nil
}
