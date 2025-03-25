package entity

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

type ListResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Count   int    `json:"count"`
	Total   int    `json:"total"`
	Results []T    `json:"data"`
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

func SendListResponse[T any](
	w http.ResponseWriter, statusCode int, message string, pagination Pagination, total int, results []T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ListResponse[T]{
		Code:    statusCode,
		Message: message,
		Offset:  pagination.Offset,
		Limit:   pagination.Limit,
		Total:   total,
		Count:   len(results),
		Results: results,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Something is broken", http.StatusInternalServerError)
	}
}
