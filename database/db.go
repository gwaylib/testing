/*
以工厂的模式构建数据库，以避免数据库被多次打开。
因database/sql本身已实现连接池，因此没有必要创建多个同一的数据库连接实例
*/
package database

import (
	"database/sql"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	gorp "gopkg.in/gorp.v2"
)

// 仅继承并重写sql.DB, 不增加新的方法，
// 以便可直接使用sql.DB的方法，提高访问效率与降低使用复杂性
type DB struct {
	*sql.DB
	driverName string
	isClose    bool
	mu         sync.Mutex
	xdb        *sqlx.DB
	orp        *gorp.DbMap
}

func newDB(drvName string, db *sql.DB) *DB {
	var dialect gorp.Dialect
	switch {
	case strings.Index(drvName, "mysql") > -1:
		dialect = gorp.MySQLDialect{}
	case strings.Index(drvName, "sqlite") > -1:
		dialect = gorp.SqliteDialect{}
	case strings.Index(drvName, "oracle") > -1, strings.Index(drvName, "oci8") > -1:
		dialect = gorp.OracleDialect{}
	case strings.Index(drvName, "postgres") > -1:
		dialect = gorp.PostgresDialect{}
	case strings.Index(drvName, "sqlserver") > -1, strings.Index(drvName, "mssql") > -1:
		dialect = gorp.SqlServerDialect{}
	default:
		panic("unsport drv:" + drvName)
	}
	return &DB{
		DB:         db,
		driverName: drvName,
		xdb:        sqlx.NewDb(db, drvName),
		orp:        &gorp.DbMap{Db: db, Dialect: dialect},
	}
}

func (db *DB) DriverName() string {
	return db.driverName
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
