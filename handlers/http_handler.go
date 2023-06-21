package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/7dpk/keyvaluestore/database"
)

// handle the HTTP requests and interact with the Database
type HTTPHandler struct {
	Database *database.Database
}

// represent the request body JSON structure
type RequestBody struct {
	Command string `json:"command"`
}

// represent the error response JSON structure
type ResponseError struct {
	Error string `json:"error"`
}

// represent the value response JSON structure
type ResponseValue struct {
	Value string `json:"value"`
}

type ResponseBlank struct{}

// write the error response JSON to the response writer
func writeErrorJSON(w http.ResponseWriter, errMsg string, statusCode int) {
	response := ResponseError{
		Error: errMsg,
	}
	writeJSONResponse(w, response, statusCode)
}

// write the value response JSON to the response writer
func writeValueJSON(w http.ResponseWriter, value string) {
	response := ResponseValue{
		Value: value,
	}
	writeJSONResponse(w, response, http.StatusOK)
}

// write the blank response JSON to the response writer
func writeBlankJSON(w http.ResponseWriter) {
	response := ResponseBlank{}
	writeJSONResponse(w, response, http.StatusOK)
}

// write the given response object as JSON to the response writer
func writeJSONResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

// handle the HTTP request and perform the appropriate database operation
func (h *HTTPHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		writeErrorJSON(w, "invalid request body", http.StatusBadRequest)
		return
	}

	command := requestBody.Command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		writeErrorJSON(w, "empty command", http.StatusBadRequest)
		return
	}

	switch parts[0] {
	case "SET":
		if len(parts) < 3 {
			writeErrorJSON(w, "invalid SET command", http.StatusBadRequest)
			return
		}
		key := parts[1]
		value := parts[2]
		var expiry time.Duration
		var condition string

		if len(parts) > 3 {
			for i := 3; i < len(parts); i++ {
				part := parts[i]
				if part == "EX" && i+1 < len(parts) {
					expiryStr := parts[i+1]
					expirySeconds, err := strconv.Atoi(expiryStr)
					if err != nil {
						writeErrorJSON(w, "invalid expiry time", http.StatusBadRequest)
						return
					}
					expiry = time.Duration(expirySeconds) * time.Second
					i++
				} else if part == "NX" || part == "XX" {
					condition = part
				}
			}
		}

		err := h.Database.Set(key, value, expiry, condition)
		if err != nil {
			writeErrorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeBlankJSON(w)
	case "GET":
		if len(parts) != 2 {
			writeErrorJSON(w, "invalid GET command", http.StatusBadRequest)
			return
		}
		key := parts[1]
		value, err := h.Database.Get(key)
		if err != nil {
			writeErrorJSON(w, err.Error(), http.StatusNotFound)
			return
		}
		writeValueJSON(w, value)
	case "QPUSH":
		if len(parts) < 3 {
			writeErrorJSON(w, "invalid QPUSH command", http.StatusBadRequest)
			return
		}
		key := parts[1]
		values := parts[2:]
		h.Database.QPush(key, values)
		writeBlankJSON(w)
	case "QPOP":
		if len(parts) != 2 {
			writeErrorJSON(w, "invalid QPOP command", http.StatusBadRequest)
			return
		}
		key := parts[1]
		value, err := h.Database.QPop(key)
		if err != nil {
			writeErrorJSON(w, err.Error(), http.StatusNotFound)
			return
		}
		writeValueJSON(w, value)
	case "BQPOP":
		if len(parts) != 3 {
			writeErrorJSON(w, "invalid BQPOP command", http.StatusBadRequest)
			return
		}
		key := parts[1]
		timeoutStr := parts[2]
		timeoutSeconds, err := strconv.ParseFloat(timeoutStr, 64)
		if err != nil {
			writeErrorJSON(w, "invalid timeout", http.StatusBadRequest)
			return
		}
		timeout := time.Duration(timeoutSeconds) * time.Second
		value, err := h.Database.BQPop(key, timeout)
		if err != nil {
			writeErrorJSON(w, err.Error(), http.StatusNotFound)
			return
		}
		writeValueJSON(w, value)
	default:
		writeErrorJSON(w, "invalid command", http.StatusBadRequest)
	}
}
