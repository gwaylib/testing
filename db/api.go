/*
此包以工厂的模式提供数据库连接，以便优化数据库连接数
*/
package db

import (
	"database/sql"
	"io"
	"sync"

	"github.com/gwaylib/datastore/conf/etc"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
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
	odb, err := sql.Open(drvName, dsn)
	if err != nil {
		return nil, errors.As(err)
	}
	db = &DB{DB: odb, isClose: false}
	cache[key] = db
	return db, nil
}

// 获取数据数连接实例
func GetDB(etcFileName, sectionName string) *DB {
	db, err := getDB(etcFileName, sectionName)
	if err != nil {
		panic(err)
	}
	return db
}

// 检查数据库是否存在并返回数据连接实例
func HasDB(etcFileName, sectionName string) (*DB, error) {
	return getDB(etcFileName, sectionName)
}

// 提供此懒的关闭方法，调用者不需要处理错误
func Close(closer io.Closer) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil {
		log.Warn(errors.As(err))
	}
}

// 提供懒处理的回滚方法，调用者不需要处理错误
func Rollback(tx *sql.Tx) {
	err := tx.Rollback()

	// roll back error is a serious error
	if err != nil {
		log.Error(errors.As(err))
	}
}
