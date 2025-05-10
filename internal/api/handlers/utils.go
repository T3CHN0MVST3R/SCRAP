package handlers

import (
	"encoding/json"
	"net/http"

	"website-scraper/internal/models"
)

// RespondWithError отправляет клиенту ошибку в формате JSON
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, models.ErrorResponse{Error: message})
}

// RespondWithJSON отправляет клиенту данные в формате JSON
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
