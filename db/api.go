/*
此包以工厂的模式提供数据库连接，以便优化数据库连接数
*/
package db

import (
	"database/sql"
	"io"

	"git.ot24.net/go/engine/errors"
	"git.ot24.net/go/engine/log"
)

// 获取数据数连接实例
func GetDB(etcFileName, secName string) *DB {
	db, err := getDB(etcFileName, secName)
	if err != nil {
		panic(err)
	}
	return db
}

// 检查数据库是否存在并返回数据连接实例
func HasDB(etcFileName, secName string) (*DB, error) {
	return getDB(etcFileName, secName)
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
