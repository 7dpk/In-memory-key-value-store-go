package handlers_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/7dpk/keyvaluestore/database"
	"github.com/7dpk/keyvaluestore/handlers"
)

func TestQPUSH_QPOP(t *testing.T) {

	db := database.NewDatabase()
	handler := &handlers.HTTPHandler{
		Database: db,
	}
	server := httptest.NewServer(http.HandlerFunc(handler.HandleRequest))
	defer server.Close()

	// Test cases for QPUSH command
	// write code which makes POST request to server with body as JSON {"command": "QPUSH key 1 2 3"} and check the response
	// reponse code should be ok and body should be empty

	requestBody := fmt.Sprintf(`{"command": "%s"}`, "QPUSH key 1 2 3")
	reader := strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPUSH key 1 2 3")

	// Send the request
	resp, err := http.Post(server.URL, "application/json", reader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	// Test cases for QPOP command
	// write code which makes POST request to server with body as JSON {"command": "QPOP key"} and check the response
	// reponse code should be ok and body should be first element of queue i.e. 3 then 2 then 1
	requestBody = fmt.Sprintf(`{"command": "%s"}`, "QPOP key")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPOP key")

	// Send the request
	resp, err = http.Post(server.URL, "application/json", reader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	var response Response

	// Decode the response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Check the response body
	if response.Value != "3" {
		t.Errorf("Expected response body to be 3; got %v", response.Value)
	}

	// Make the request again this time value should be 2
	requestBody = fmt.Sprintf(`{"command": "%s"}`, "QPOP key")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPOP key")

	// Send the request
	resp, err = http.Post(server.URL, "application/json", reader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	// Decode the response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Check the response body

	if response.Value != "2" {
		t.Errorf("Expected response body to be 2; got %v", response.Value)
	}

	// Make the request again this time value should be 1

	requestBody = fmt.Sprintf(`{"command": "%s"}`, "QPOP key")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPOP key")

	// Send the request
	resp, err = http.Post(server.URL, "application/json", reader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	// Decode the response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Check the response body
	if response.Value != "1" {
		t.Errorf("Expected response body to be 1; got %v", response.Value)
	}

	// Make the request again this time we should get an error "key not found" and status should be 404

	requestBody = fmt.Sprintf(`{"command": "%s"}`, "QPOP key")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPOP key")

	// Send the request
	resp, err = http.Post(server.URL, "application/json", reader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status BadRequest; got %v", resp.StatusCode)
	}

	// Decode the response

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Check the response body

	if response.Error != "queue is empty" {
		t.Errorf("Expected response body to be \"key not found\"; got %v", response.Error)
	}

}
