package httptools

import (
	"encoding/json"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSONResponse(w, statusCode, map[string]string{"error": message})
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
