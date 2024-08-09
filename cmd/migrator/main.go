package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		host            string
		port            int
		user            string
		password        string
		dbname          string
		sslmode         string
		migrationsPath  string
		migrationsTable string
	)

	// Парсинг флагов
	flag.StringVar(&host, "host", "localhost", "database host")
	flag.IntVar(&port, "port", 5432, "database port")
	flag.StringVar(&user, "user", "postgres", "database user")
	flag.StringVar(&password, "password", "", "database password")
	flag.StringVar(&dbname, "dbname", "calendar", "database name")
	flag.StringVar(&sslmode, "sslmode", "disable", "ssl mode")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if migrationsPath == "" {
		log.Fatal("storage-path is required")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		user, password, host, port, dbname, sslmode, migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)

	if err != nil {
		log.Fatalf("Failed to create migrate instanse: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		log.Fatalf("Failed to apply migrations: %v", err)
	}

	fmt.Println("migrations applied successfully")
}
