/*
此包以工厂的模式提供数据库连接，以便优化数据库连接数
*/
package database

import (
	"database/sql"
	"io"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

// 返回一个全新的实例
func Open(drvName, dsn string) (*DB, error) {
	db, err := sql.Open(drvName, dsn)
	if err != nil {
		return nil, errors.As(err, drvName, dsn)
	}
	return &DB{DB: db, driverName: drvName}, nil
}

// 获取数据库池中的实例
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
