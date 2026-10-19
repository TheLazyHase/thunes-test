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

func GetDevDB() (*sqlx.DB, error) {
	db, err := GetDB("postgres://postgres:postgres@db:5432/thunes-db?sslmode=disable")
	if err != nil {
		return db, fmt.Errorf("development DB retrieval: %w", err)
	}
	err = MigrateDB(db, "file://migrations")
	if err != nil {
		return db, fmt.Errorf("development DB migration: %w", err)
	}
	SeedDevDB(db)
	return db, nil
}

func GetLocalDB() (*sqlx.DB, error) {
	db, err := GetDB("postgres://postgres:postgres@0.0.0.0:6432/thunes-db?sslmode=disable")
	if err != nil {
		return db, fmt.Errorf("local DB retrieval: %w", err)
	}
	SeedDevDB(db)
	return db, nil
}

func GetDB(addr string) (*sqlx.DB, error) {

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

	return db, nil
}

func MigrateDB(db *sqlx.DB, migration string) error {
	instance, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		migration,
		"thunes-db",
		instance,
	)
	if err != nil {
		return fmt.Errorf("Error migrating the DB: %s", err.Error())
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Error changing DB version", err)
	}
	return nil
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
