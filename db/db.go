/*
以工厂的模式构建数据库，以避免数据库被多次打开。
因database/sql本身已实现连接池，因此没有必要创建多个同一的数据库连接实例
*/
package db

import (
	"database/sql"
	"sync"

	"github.com/gwaylib/datastore/conf/etc"
	"github.com/gwaylib/errors"
)

// 仅继承并重写sql.DB, 不增加新的方法，
// 以便可直接使用sql.DB的方法，提高访问效率与隆低使用复杂性
type DB struct {
	*sql.DB
	isClose bool
	mu      sync.Mutex
}

func (db *DB) IsClose() bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.isClose
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.isClose = true
	return db.DB.Close()
}

var cache = sync.Map{}

func getDB(etcFileName, sectionName string) (*DB, error) {
	idb, ok := cache.Load(etcFileName + sectionName)
	if ok {
		db := idb.(*DB)
		if db != nil && !db.isClose {
			return db, nil
		}
	}

	// open
	cfg, err := etc.GetEtc(etcFileName)
	if err != nil {
		return nil, errors.As(err, etcFileName)
	}
	drvName := cfg.Section(sectionName).Key("driver").String()
	dsn := cfg.Section(secName).Key("dsn").String()
	odb, err := sql.Open(drvName, dsn)
	if err != nil {
		return nil, errors.As(err)
	}
	db := &DB{DB: odb, isClose: false}
	cache.Store(etcFileName+sectionName, db)
	return db, nil
}
