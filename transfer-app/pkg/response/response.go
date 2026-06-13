package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, Response{Success: true, Message: message, Data: data})
}

func Created(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, Response{Success: false, Message: "Bad request", Error: message})
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, Response{Success: false, Message: "Unauthorized", Error: message})
}

func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, Response{Success: false, Message: "Forbidden", Error: message})
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, Response{Success: false, Message: "Not found", Error: message})
}

func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, Response{Success: false, Message: "Conflict", Error: message})
}

func UnprocessableEntity(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnprocessableEntity, Response{Success: false, Message: "Unprocessable entity", Error: message})
}

func InternalServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, Response{Success: false, Message: "Internal server error", Error: message})
}

func WithPagination(w http.ResponseWriter, message string, data interface{}, meta Meta) {
	JSON(w, http.StatusOK, Response{Success: true, Message: message, Data: data, Meta: &meta})
}
