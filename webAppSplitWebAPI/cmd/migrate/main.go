package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {

	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal(envErr)
	}

	dbDsn := os.Getenv("DB_DSN")

	m, err := migrate.New("file://migrations", dbDsn)

	if err != nil {

		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {

		log.Fatal(err)
	}
}
