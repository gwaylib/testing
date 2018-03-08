package database

import "github.com/jmoiron/sqlx"

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.xdb.Select(dest, query, args...)
}

func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.xdb.Get(dest, query, args...)
}

func (db *DB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.xdb.Queryx(query, args...)
}

func (db *DB) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return db.xdb.QueryRowx(query, args...)
}

func (db *DB) Beginx() (*sqlx.Tx, error) {
	return db.xdb.Beginx()
}

func (db *DB) Preparex(query string) (*sqlx.Stmt, error) {
	return db.xdb.Preparex(query)
}

func (db *DB) Sqlx() *sqlx.DB {
	return db.xdb
}
