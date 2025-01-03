package main

import (
	"database/sql"
	"fmt"
	"os"

	"gurch101.github.io/go-web/pkg/dbutils"
)

func main() {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=1&_journal=WAL", os.Getenv("DB_FILEPATH")))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = dbutils.SeedDb(db)
	if err != nil {
		panic(err)
	}
}
