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

const (
	DRV_NAME_MYSQL     = "msyql"
	DRV_NAME_ORACLE    = "oracle" // or "oic8" is supported
	DRV_NAME_POSTGRES  = "postgres"
	DRV_NAME_SQLITE3   = "sqlite3"
	DRV_NAME_SQLSERVER = "sqlserver" // or "mssql" is supported
)

var (
	DEFAULT_DRV_NAME = DRV_NAME_MYSQL
)

// 使用一个已有的标准数据库实例构建出实例
func NewDB(drvName string, db *sql.DB) *DB {
	return newDB(drvName, db)
}

// 返回一个全新的实例
func Open(drvName, dsn string) (*DB, error) {
	db, err := sql.Open(drvName, dsn)
	if err != nil {
		return nil, errors.As(err, drvName, dsn)
	}
	return newDB(drvName, db), nil
}

// 注册一个池实例
func RegCache(iniFileName, sectionName string, db *DB) {
	regCache(iniFileName, sectionName, db)
}

// 获取数据库池中的实例
// 如果不存在，会使用配置文件进行读取
func CacheDB(iniFileName, sectionName string) *DB {
	db, err := cacheDB(iniFileName, sectionName)
	if err != nil {
		panic(err)
	}
	return db
}

// 检查数据库是否存在并返回数据连接实例
func HasDB(etcFileName, sectionName string) (*DB, error) {
	return cacheDB(etcFileName, sectionName)
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

// 添加一条数据，需要结构体至少标注字段名 `db:"name"`, 标签详情请参考github.com/jmoiron/sqlx
// 关于drvNames的设计说明
// 因支持一个可变参数, 或未填，将使用默认值:DEFAULT_DRV_NAME
func InsertStruct(exec Execer, obj interface{}, tbName string, drvNames ...string) (sql.Result, error) {
	return insertStruct(exec, obj, tbName, drvNames...)
}

// 扫描结果到一个结构体，该结构体可以是数组
// 代码设计请参阅github.com/jmoiron/sqlx
func ScanStruct(rows Scaner, obj interface{}) error {
	return scanStruct(rows, obj)
}

// 查询一个对象
func QueryObj(db Queryer, obj interface{}, querySql string, args ...interface{}) error {
	return queryObj(db, obj, querySql, args...)
}

// 执行一个通用的数字查询
func QueryInt(db Queryer, querySql string, args ...interface{}) (int64, error) {
	return queryInt(db, querySql, args...)
}

// 执行一个通用的字符查询
func QueryStr(db Queryer, querySql string, args ...interface{}) (string, error) {
	return queryStr(db, querySql, args...)
}

// 执行一个通用的查询
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func QueryTable(db Queryer, querySql string, args ...interface{}) (titles []string, result [][]interface{}, err error) {
	return queryTable(db, querySql, args...)
}

// 查询一条数据，并发map结构返回，以便页面可以直接调用
// 因需要查标题，相对标准sql会慢一些，适用于偷懒查询的方式
// 即使发生错误返回至少是零长度的值
func QueryMap(db Queryer, querySql string, args ...interface{}) ([]map[string]interface{}, error) {
	return queryMap(db, querySql, args...)
}
