package pg

import (
	"github.com/gwaylib/datastore/database"
	_ "github.com/lib/pq"
)

func Open(drvName, dsn string) (*database.DB, error) {
	return database.Open(drvName, dsn)
}

func GetDB(etcFileName, secName string) *database.DB {
	return database.GetDB(etcFileName, secName)
}
