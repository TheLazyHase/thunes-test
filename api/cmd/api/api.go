package main

import (
	"log"
	"net/http"

	"thunes-api/internal/handlers"
	"thunes-api/internal/users"
	"thunes-api/internal"

	"github.com/gorilla/mux"
)

func main() {
	db, err := internal.GetLocalDB()

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
