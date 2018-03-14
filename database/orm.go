package database

import (
	"github.com/jmoiron/sqlx"
	"gopkg.in/gorp.v2"
)

// gorp ORM框架
// https://github.com/go-gorp/gorp
func (db *DB) Gorp() *gorp.DbMap {
	return db.orp
}

// sqlx ORM框架
// https://github.com/jmoiron/sqlx
func (db *DB) Sqlx() *sqlx.DB {
	return db.xdb
}
