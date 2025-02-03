package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := TestStruct{Name: "John", Age: 30}

	err := writeJSON(rr, http.StatusOK, data)
	if err != nil {
		t.Fatalf("writeJSON returned an error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response TestStruct
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if response.Name != "John" || response.Age != 30 {
		t.Errorf("Unexpected JSON response: %+v", response)
	}
}

func TestReadJSON(t *testing.T) {

	t.Run("should parse json", func(t *testing.T) {
		validJson := `{"name": "John", "age": 30}`

		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(validJson)))
		rr := httptest.NewRecorder()

		var result TestStruct
		err := readJSON(rr, req, &result)
		if err != nil {
			t.Fatalf("readJSON returned an error: %v", err)
		}

		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected {John, 30}, got %+v", result)
		}
	})

	t.Run("should return error for invalid json", func(t *testing.T) {
		validJson := `{"name": "John", "age":}`

		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(validJson)))
		rr := httptest.NewRecorder()

		var result TestStruct
		err := readJSON(rr, req, &result)
		if err == nil {
			t.Fatalf("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("should return error unknown json field", func(t *testing.T) {
		validJson := `{"name": "John", "time":"12:00:00"}`

		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(validJson)))
		rr := httptest.NewRecorder()

		var result TestStruct
		err := readJSON(rr, req, &result)
		if err == nil {
			t.Fatalf("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("should return error too large json", func(t *testing.T) {
		largeJSON := `{"name":"` + string(make([]byte, 1_048_579)) + `"}`
		req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(largeJSON)))
		rr := httptest.NewRecorder()

		var result TestStruct
		err := readJSON(rr, req, &result)
		if err == nil {
			t.Fatalf("Expected error for invalid JSON, got nil")
		}
	})
}

func TestWriteJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	err := writeJSONError(rr, http.StatusBadRequest, "test error")

	if err != nil {
		t.Fatalf("writeJSONError returned an error: %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	expected := "test error"
	if response["error"] != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, response["error"])
	}
}

func TestJsonRespone(t *testing.T) {
	rr := httptest.NewRecorder()

	var result TestStruct
	result.Name = "John"
	result.Age = 30
	app := &application{}
	err := app.jsonResponse(rr, http.StatusOK, result)

	if err != nil {
		t.Fatalf("jsonResponse returned an error: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	if !strings.Contains(rr.Body.String(), `"name":"John"`) {
		t.Errorf("Unexpected JSON response: %s", rr.Body.String())
	}
}
