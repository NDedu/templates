package main

import (
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {

	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal(envErr)
	}

	dbDsn := os.Getenv("DB_DSN")

	migrateDsn := dbDsn
	if !strings.HasPrefix(migrateDsn, "sqlite3://") {

		migrateDsn = "sqlite3://" + migrateDsn
	}

	m, err := migrate.New("file://migrations", migrateDsn)

	if err != nil {

		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {

		log.Fatal(err)
	}
}
