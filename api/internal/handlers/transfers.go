package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"

	"thunes-api/internal/users"
	"thunes-api/internal/transfers"
)

type transfersHandler struct {
	service transfers.Service
	userService users.Service
}

func NewTransfersHandler(s transfers.Service, u users.Service) (*transfersHandler, error) {
	return &transfersHandler{
		service: s,
		userService: u,
	}, nil
}


type createTransfersRequest struct {
	Destination int64 `json:"destination_id" binding:"required"`
	Amount      int64 `json:"amount" binding:"required"`
}

func (h transfersHandler) GetTransfers(writer http.ResponseWriter, request *http.Request) {
	user := request.Context().Value("user")
	if user == nil {
		log.Printf("No user found")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	transfers, err := h.service.GetTransfers(request.Context(), *user.(*users.User))
	if err != nil {
		log.Printf("Problem getting transfers: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(transfers)
	if err != nil {
		log.Printf("Error marshaling value: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(result)
}

func (h transfersHandler) CreateTransfer(writer http.ResponseWriter, request *http.Request) {
	reqDump, _ := httputil.DumpRequest(request, true)
	log.Printf("REQUEST:\n%s", string(reqDump))

	decoder := json.NewDecoder(request.Body)

	var req createTransfersRequest
	err := decoder.Decode(&req)

	if err != nil {
		log.Printf("Error decoding request: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := request.Context().Value("user")
	if user == nil {
		log.Printf("No user found")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	destination, err := h.userService.GetUser(request.Context(), req.Destination)
	if err != nil {
		log.Printf("Destination not found")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	transfer, errorCode, err := h.service.CreateTransfer(request.Context(), *user.(*users.User), *destination, req.Amount)
	if err != nil {
		log.Printf("Problem creating transfers: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		if errorCode != "" {
			writer.Write([]byte(errorCode))
		}
		return
	}
	result, err := json.Marshal(transfer)
	//TODO: should we rollback if we cannot marshal the result ?
	if err != nil {
		log.Printf("Error marshaling value: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(result)
}
