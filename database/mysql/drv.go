package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gwaylib/datastore/database"
)

func Open(drvName, dsn string) (*database.DB, error) {
	return database.Open(drvName, dsn)
}

func GetDB(etcFileName, secName string) *database.DB {
	return database.GetDB(etcFileName, secName)
}
