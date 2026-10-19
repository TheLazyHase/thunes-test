package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"thunes-api/internal/users"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type userHandler struct {
	service users.Service
}

func NewUserHandler(s users.Service) (*userHandler, error) {
	return &userHandler{
		service: s,
	}, nil
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var jwtKey = []byte("secret_key")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (h userHandler) Login(writer http.ResponseWriter, request *http.Request) {
	reqDump, _ := httputil.DumpRequest(request, true)
	log.Printf("REQUEST:\n%s", string(reqDump))

	decoder := json.NewDecoder(request.Body)

	var req loginRequest
	err := decoder.Decode(&req)

	if err != nil {
		log.Printf("Error decoding request: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.service.Login(request.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("Login issue: %s", err)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Unauthorized"))
		return
	}

	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error generating JWT token %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Write([]byte(tokenString))
	writer.WriteHeader(http.StatusOK)

	log.Printf("User %s succesfully logged in", req.Username)
	log.Printf("Generated JWT Token: %s", tokenString)
}
