package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/7dpk/keyvaluestore/database"
	"github.com/7dpk/keyvaluestore/handlers"
)

type Response struct {
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type Request struct {
	Command string `json:"command"`
}

func TestSetAndGetCommands(t *testing.T) {
	// Initialize the server and database
	db := database.NewDatabase()
	handler := &handlers.HTTPHandler{
		Database: db,
	}
	server := httptest.NewServer(http.HandlerFunc(handler.HandleRequest))
	defer server.Close()

	// Test cases for SET command
	setTestCases := []struct {
		Command  string
		Status   int
		ExpValue string
		ExpError bool
		ErrValue string
	}{
		// Valid SET command without expiry and condition
		{
			Command:  `SET hello world`,
			Status:   http.StatusOK,
			ExpValue: "world",
			ExpError: false,
			ErrValue: "",
		},
		// Valid SET command with condition NX for non-existing key
		{
			Command:  `SET hello key NX`,
			Status:   http.StatusBadRequest,
			ExpValue: "",
			ExpError: true,
			ErrValue: "key already exists",
		},
		// Valid SET command with condition XX for existing key
		{
			Command:  `SET hello updated XX`,
			Status:   http.StatusOK,
			ExpValue: "updated",
			ExpError: false,
			ErrValue: "",
		},
		// Valid SET command with expiry
		{
			Command:  `SET foo bar EX 1`,
			Status:   http.StatusOK,
			ExpValue: "bar",
			ExpError: false,
			ErrValue: "",
		},
	}

	for _, testCase := range setTestCases {
		// Construct the request body
		requestBody := fmt.Sprintf(`{"command": "%s"}`, testCase.Command)
		reader := strings.NewReader(requestBody)

		log.Println("Testing command: " + testCase.Command)

		// Send the request
		resp, err := http.Post(server.URL, "application/json", reader)
		if err != nil {
			t.Errorf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode != testCase.Status {
			t.Errorf("Expected status code %d, but got %d", testCase.Status, resp.StatusCode)
		}

		// Make the GET requests to check the values if there is any update
		if !testCase.ExpError {
			getCommand := fmt.Sprintf("GET %v", strings.Split(testCase.Command, " ")[1])
			data := Request{
				Command: getCommand,
			}

			body, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling JSON:", err)
				return
			}
			get_resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}
			defer get_resp.Body.Close()
			// Check the value if error doesn't exist
			if get_resp.StatusCode != 200 {
				t.Errorf("Expected statusCode %v, but got %v", 200, get_resp.StatusCode)
			} else {
				responseBody, err := io.ReadAll(get_resp.Body)
				if err != nil {
					fmt.Printf("Error reading body: %v", err)
				}
				var response Response
				err = json.Unmarshal(responseBody, &response)
				if err != nil {
					fmt.Printf("Failed to parse response JSON: %v", err)
				}
				if !testCase.ExpError {
					// value, _ := db.Get(strings.Split(testCase.Command, " ")[1])
					if response.Value != testCase.ExpValue {
						t.Errorf("Expected value %s, but got %s", testCase.ExpValue, response.Value)
					}
				} else {
					log.Println("Error expected: key not found" + " Got: " + err.Error())
					if err != errors.New(testCase.ErrValue) {
						t.Errorf("Expected err %v, but got %v", testCase.ExpError, err)
					}

				}
			}
		}
	}

	// Test cases for GET command
	getTestCases := []struct {
		Key      string
		Expected string
		ExpError bool
		ErrValue string
		WaitTime int
		Status   int
	}{
		// Valid GET command for updated key
		{
			Key:      "hello",
			Expected: "updated",
			WaitTime: 0,
			ExpError: false,
			ErrValue: "",
			Status:   http.StatusOK,
		},
		// Valid GET command for non-existing key
		{
			Key:      "nonexisting",
			Expected: "",
			WaitTime: 0,
			ExpError: true,
			ErrValue: "key not found",
			Status:   http.StatusNotFound,
		},
		// Valid Get command for expired key
		{
			Key:      "foo",
			Expected: "",
			WaitTime: 2,
			ExpError: true,
			ErrValue: "key not found",
			Status:   http.StatusNotFound,
		},
	}

	for _, testCase := range getTestCases {
		// wait for WaitTime for the key to be expired
		if testCase.WaitTime != 0 {
			log.Println("Waiting " + fmt.Sprint(testCase.WaitTime) + " Seconds" + " for key: " + testCase.Key + " to expire")
		}
		time.Sleep(time.Second * time.Duration(testCase.WaitTime))

		// Construct the request Body
		getCommand := fmt.Sprintf("GET %v", testCase.Key)
		data := Request{
			Command: getCommand,
		}
		body, err := json.Marshal(data)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			return
		}
		// Send the request
		log.Println("Test command: " + getCommand)

		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		var response Response

		err = json.Unmarshal(responseBody, &response)
		if testCase.ExpError {
			if response.Error != testCase.ErrValue {
				t.Errorf("Expected error to be %q found %q", testCase.ErrValue, response.Error)
			}
		} else {
			if response.Value != testCase.Expected {
				t.Errorf("Expected value to be %q found %q", testCase.Expected, response.Value)
			}
		}
	}
}
