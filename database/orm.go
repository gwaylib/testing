package database

import (
	"strings"

	"gopkg.in/gorp.v2"
)

// TODO: 选型ORM, 除gorp外，其他主流框架数据库打开方式有些冲突，暂不做支持。

// gorp ORM框架
// https://github.com/go-gorp/gorp
func (db *DB) GORP() *gorp.DbMap {
	var dialect gorp.Dialect
	switch {
	case strings.Index(db.driverName, "mysql") > -1:
		dialect = gorp.MySQLDialect{}
	case strings.Index(db.driverName, "sqlite") > -1:
		dialect = gorp.SqliteDialect{}
	case strings.Index(db.driverName, "oracle") > -1, strings.Index(db.driverName, "oci8") > -1:
		dialect = gorp.OracleDialect{}
	case strings.Index(db.driverName, "postgres") > -1:
		dialect = gorp.PostgresDialect{}
	case strings.Index(db.driverName, "sqlserver") > -1, strings.Index(db.driverName, "mssql") > -1:
		dialect = gorp.SqlServerDialect{}
	default:
		panic("unsport driver:" + db.driverName)
	}
	// TODO: 优化实例生成
	return &gorp.DbMap{Db: db.DB, Dialect: dialect}
}
