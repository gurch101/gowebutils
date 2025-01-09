package main

import (
	"database/sql"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

func main() {
	dbDSN := parser.ParseEnvStringPanic("DB_FILEPATH") + "?_foreign_keys=1&_journal=WAL"

	db, err := sql.Open(dbutils.SqliteDriverName, dbDSN)
	if err != nil {
		panic(err)
	}

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	err = dbutils.SeedDB(db)
	if err != nil {
		panic(err)
	}
}
