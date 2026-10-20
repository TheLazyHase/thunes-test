package internal

import (
	"fmt"
	"log"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func GetTestDB() (*sqlx.DB, error) {
	return GetDB("postgres://postgres:postgres@0.0.0.0:6432/thunes-db?sslmode=disable", "file://migrations", "dev")
}

func GetLocalDB() (*sqlx.DB, error) {
	return GetDB("postgres://postgres:postgres@db:5432/thunes-db?sslmode=disable", "file://migrations", "dev")
}

func GetDB(addr string, migration string, environment string) (*sqlx.DB, error) {
	//TODO: make a proper environment-dependant seed
	if environment != "dev" {
		return nil, fmt.Errorf("Unrecognized environment: %s", environment)
	}

	connConfig, err := pgx.ParseConfig(addr)
	if err != nil {
		errMsg := err.Error()
		errMsg = regexp.MustCompile(`(://[^:]+:).+(@.+)`).ReplaceAllString(errMsg, "$1*****$2")
		errMsg = regexp.MustCompile(`(password=).+(\s+)`).ReplaceAllString(errMsg, "$1*****$2")
		return nil, fmt.Errorf("parsing DSN failed: %s", errMsg)
	}
	connectionStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sqlx.Open("pgx", connectionStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	instance, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		migration,
		"thunes-db",
		instance,
	)
	if err != nil {
		return nil, fmt.Errorf("Error migrating the DB: %s", err.Error())
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("Error changing DB version", err)
	}

	SeedDevDB(db)

	return db, nil
}

func SeedDevDB(db *sqlx.DB) {
	users := []struct {
		Username string
		Password string
	}{
		{Username: "Alice", Password: "password123"},
		{Username: "Bob", Password: "password123"},
		{Username: "Charlie", Password: "password123"},
		{Username: "David", Password: "password123"},
	}
	for _, user := range users {

		db.Exec("INSERT INTO users(username, password) VALUES ($1,$2) ON CONFLICT DO NOTHING", user.Username, user.Password)
	}

	log.Println("Database Seeding Completed.")
}
