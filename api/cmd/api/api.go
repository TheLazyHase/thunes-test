package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"thunes-api/internal/handlers"
	"thunes-api/internal/transfers"
	"thunes-api/internal/users"
	"thunes-api/internal"

	"github.com/gorilla/mux"
	"github.com/dgrijalva/jwt-go"
)

type authenticationMiddleware struct {
	Service users.Service
}

func main() {
	//TODO: deal with environments
	db, err := internal.GetDevDB()

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

	transferRepo, err := transfers.NewRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	transferService, err := transfers.NewService(transferRepo, userRepo)
	if err != nil {
		log.Fatal(err)
	}
	transferHandler, err := handlers.NewTransfersHandler(transferService, userService)
	if err != nil {
		log.Fatal(err)
	}

	amw := authenticationMiddleware{
		Service: userService,
	}

	router := mux.NewRouter()
	router.Use(logRequests)
	router.Use(setAccessControlHeader)
	router.Use(amw.authMiddleware)

	router.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/me", userHandler.Me).Methods(http.MethodGet)
	router.HandleFunc("/users", userHandler.ListUsers).Methods(http.MethodGet)
	router.HandleFunc("/transfers", transferHandler.CreateTransfer).Methods(http.MethodPost)
	router.HandleFunc("/transfers", transferHandler.GetTransfers).Methods(http.MethodGet)
	router.Methods(http.MethodOptions).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}).Methods(http.MethodOptions)

	log.Println("Application has started. Listening port is 4000")
	http.ListenAndServe(":4000", router)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqDump, _ := httputil.DumpRequest(r, true)
		log.Printf("REQUEST:\n%s", string(reqDump))
		// Extract claims and pass to next handler
		next.ServeHTTP(w, r)
	})
}

func setAccessControlHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(w, r)
	})
}

func (amw *authenticationMiddleware) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//No auth required for login
		if strings.TrimPrefix(r.URL.Path, "/") == "login" {
			next.ServeHTTP(w, r)
			return
		}
		//Neither for OPTION
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == "" {
			log.Printf("No token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.ParseWithClaims(tokenStr, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return handlers.JWTKey, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims := token.Claims.(*handlers.Claims)
		log.Printf("token have %s and %d", claims.Username, claims.ID)
		currentUser, err := amw.Service.GetUser(r.Context(), claims.ID)
		if err != nil  {
			log.Printf("Error retrieving user: %s", err.Error())
		}
		ctx := context.WithValue(r.Context(), "user", currentUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
