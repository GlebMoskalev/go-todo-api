package response

import (
	"encoding/json"
	"net/http"
)

const ServerFailureMessage = "Something went wrong, please try again later"

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type Data struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Total   int   `json:"total"`
	Count   int   `json:"count"`
	Results []any `json:"results"`
}

func SendResponse[T any](w http.ResponseWriter, statusCode int, message string, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := Response[T]{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Something is broken", http.StatusInternalServerError)
	}
}
