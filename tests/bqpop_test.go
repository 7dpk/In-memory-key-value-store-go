package handlers_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/7dpk/keyvaluestore/database"
	"github.com/7dpk/keyvaluestore/handlers"
)

func TestBQPOP(t *testing.T) {

	db := database.NewDatabase()
	handler := &handlers.HTTPHandler{
		Database: db,
	}
	server := httptest.NewServer(http.HandlerFunc(handler.HandleRequest))
	defer server.Close()

	// Test cases for BQPOP command
	// write code which makes POST request to server with body as JSON {"command": "BQPOP key 4"}
	// where 4 is timeout in seconds, if the element is not available in queue then it should wait for 4 seconds
	// and then return the element as soon it's available if available, if not available then it should return empty response
	// reponse code should be ok and body should be empty after waiting for timeout period

	// first push one element in queue
	requestBody := fmt.Sprintf(`{"command": "%s"}`, "QPUSH key 4")
	reader := strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "QPUSH key 4")

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

	// now pop the element from queue
	requestBody = fmt.Sprintf(`{"command": "%s"}`, "BQPOP key 4")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "BQPOP key 4")

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

	// Decode the JSON response
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Could not decode JSON response: %v", err)
	}

	// Check the response body as it should immediately return
	if response.Value != "4" {
		t.Errorf("Expected an 4; got %q", response.Value)
	}

	// now we run a goroutine which will BQPOP the element from queue with a timeout of 4 seconds
	// and then we will push one element in queue after 1 second of starting goroutine as soon as the element
	// is pushed in queue goroutine should return the element

	// first make a channel to receive the element from goroutine
	ch := make(chan string)

	// first start the goroutine to BQPOP the element from queue
	go func() {
		popRequest := `{"command": "BQPOP key 4"}`
		popReader := strings.NewReader(popRequest)

		log.Printf("Testing command: %q ", "BQPOP key 4")
		log.Printf("Waiting for 4 seconds before someone pushes in queue")
		// Send the request
		popResp, err := http.Post(server.URL, "application/json", popReader)
		if err != nil {
			t.Errorf("Failed to send request: %v", err)
		}
		defer popResp.Body.Close()

		// Check the response status code
		if popResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK; got %v", popResp.Status)
		}

		// Decode the JSON response
		var popResponse Response

		if err := json.NewDecoder(popResp.Body).Decode(&popResponse); err != nil {
			t.Errorf("Could not decode JSON response: %v", err)
		}

		// Check the response body as it should immediately return
		log.Printf("Response received after someone pushed in: %q ", popResponse.Value)
		if popResponse.Value != "5" {
			t.Errorf("Expected an 5; got %q", popResponse.Value)
		}

		// send the element to channel
		ch <- popResponse.Value
	}()

	// now push one element in queue after 1 second of starting goroutine
	time.Sleep(1 * time.Second)
	// now as soon the element is pushed in queue goroutine should return the element
	pushRequest := `{"command": "QPUSH key 5"}`
	pushReader := strings.NewReader(pushRequest)

	log.Printf("Testing command: %q ", "QPUSH key 5")

	// Send the request
	pushResp, err := http.Post(server.URL, "application/json", pushReader)
	if err != nil {
		t.Errorf("Failed to send request: %v", err)
	}
	defer pushResp.Body.Close()

	// Check the response status code
	if pushResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", pushResp.Status)
	}
	// check the channel for the element
	log.Printf("Waiting for element from channel")
	element := <-ch
	log.Printf("Element received from channel: %q ", element)

	// make a BQPOP request with timeout of 3 seconds this time it should timeout after 3 seconds
	// as there is no element in queue
	requestBody = fmt.Sprintf(`{"command": "%s"}`, "BQPOP key 3")
	reader = strings.NewReader(requestBody)

	log.Printf("Testing command: %q ", "BQPOP key 3")

	// Start the timer
	start := time.Now()
	log.Println("Should time out after 3 seconds")
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

	// Decode the JSON response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Could not decode JSON response: %v", err)
	}

	// Check the response body as it should timeout after 3 seconds
	if response.Value != "" {
		t.Errorf("Expected an empty string; got %q", response.Value)
	}

	// Check the duration it took to get a response
	duration := time.Since(start)
	if duration < 3*time.Second {
		t.Errorf("Expected a timeout after 3 seconds; got %v", duration)
	}
	// print the duration

	log.Printf("Duration: %v", duration)
}
