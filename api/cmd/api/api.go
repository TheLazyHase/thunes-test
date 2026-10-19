package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"thunes-api/internal/handlers"
	"thunes-api/internal/users"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := getDb()

	if err != nil {
		log.Fatal(err)
	}

	userRepo, err := users.NewRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	userService, err := users.NewService(userRepo)
	if err != nil {
		log.Fatal(err)
	}
	userHandler, err := handlers.NewUserHandler(userService)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.Use(setAccessControlHeader)

	router.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)
	router.Methods(http.MethodOptions).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)

	log.Println("Application has started. Listening port is 4000")
	http.ListenAndServe(":4000", router)
}

func setAccessControlHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(w, r)
	})
}

func getDb() (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig("postgres://postgres:postgres@db:5432/thunes-db?sslmode=disable")
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
		"file://migrations",
		"thunes-db",
		instance,
	)
	if err != nil {
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	seed(db)

	return db, nil
}

func seed(db *sqlx.DB) {
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
