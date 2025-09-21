package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
)

func TestMakeHTTPRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	client := &http.Client{}
	config := HTTPRequestConfig{
		URL:    server.URL,
		Method: "GET",
	}

	resp, err := MakeHTTPRequest(config, client)
	if err != nil {
		t.Fatalf("MakeHTTPRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestMakeJSONRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	client := &http.Client{}
	payload := map[string]string{"test": "data"}

	resp, err := MakeJSONRequest(server.URL, "POST", payload, nil, client)
	if err != nil {
		t.Fatalf("MakeJSONRequest failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestExecuteBatchStoreOperations(t *testing.T) {
	// Setup test environment
	cleanup, err := store.SetupTestPathPrefix()
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	defer cleanup()

	// Initialize store
	if err := store.Init(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Test batch operations
	ops := BatchStoreOperations{
		Sets: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Deletes: []string{}, // We'll test deletes after setting
	}

	err = ExecuteBatchStoreOperations(ops)
	if err != nil {
		t.Fatalf("ExecuteBatchStoreOperations failed: %v", err)
	}

	// Verify sets
	value1, err := store.Get("key1")
	if err != nil {
		t.Errorf("Failed to get key1: %v", err)
	}
	if value1 != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value1)
	}

	// Test deletes
	ops.Sets = map[string]string{}
	ops.Deletes = []string{"key1"}

	err = ExecuteBatchStoreOperations(ops)
	if err != nil {
		t.Fatalf("ExecuteBatchStoreOperations delete failed: %v", err)
	}

	// Verify delete
	_, err = store.Get("key1")
	if err == nil {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCheckAndReadFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	result := CheckAndReadFile(testFile)
	if !result.Exists {
		t.Error("Expected file to exist")
	}
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if string(result.Data) != testContent {
		t.Errorf("Expected '%s', got '%s'", testContent, result.Data)
	}

	// Test non-existing file
	result = CheckAndReadFile(filepath.Join(tempDir, "nonexistent.txt"))
	if result.Exists {
		t.Error("Expected file to not exist")
	}
	if result.Data != nil {
		t.Error("Expected no data for non-existent file")
	}
}

func TestSetCommonJioTVHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	deviceID := "test-device"
	crmID := "test-crm"
	uniqueID := "test-unique"

	SetCommonJioTVHeaders(req, deviceID, crmID, uniqueID)

	// Verify some key headers
	if req.Header.Get("deviceId") != deviceID {
		t.Errorf("Expected deviceId '%s', got '%s'", deviceID, req.Header.Get("deviceId"))
	}

	if req.Header.Get("crmid") != crmID {
		t.Errorf("Expected crmid '%s', got '%s'", crmID, req.Header.Get("crmid"))
	}

	if req.Header.Get("uniqueId") != uniqueID {
		t.Errorf("Expected uniqueId '%s', got '%s'", uniqueID, req.Header.Get("uniqueId"))
	}

	if req.Header.Get("appkey") != "NzNiMDhlYzQyNjJm" {
		t.Error("Expected appkey to be set")
	}
}

func TestParseJSONResponse(t *testing.T) {
	// Create test server with OK response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "test", "value": 123}`))
	}))
	defer server.Close()

	client := &http.Client{}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var target struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err = ParseJSONResponse(resp, &target)
	if err != nil {
		t.Fatalf("ParseJSONResponse failed: %v", err)
	}

	if target.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", target.Name)
	}

	if target.Value != 123 {
		t.Errorf("Expected value 123, got %d", target.Value)
	}

	// Test error response
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer errorServer.Close()

	errorResp, err := client.Get(errorServer.URL)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer errorResp.Body.Close()

	err = ParseJSONResponse(errorResp, &target)
	if err == nil {
		t.Error("Expected error for bad status code")
	}
}

func TestLogAndReturnError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	context := "test context"

	resultErr := LogAndReturnError(originalErr, context)

	if resultErr == nil {
		t.Error("Expected error to be returned")
	}

	errorMsg := resultErr.Error()
	if !contains(errorMsg, context) {
		t.Errorf("Expected error message to contain context '%s', got: %s", context, errorMsg)
	}

	if !contains(errorMsg, "original error") {
		t.Errorf("Expected error message to contain original error, got: %s", errorMsg)
	}
}

func TestSafeLogf(t *testing.T) {
	// Test with nil logger (should not crash)
	originalLog := Log
	Log = nil
	
	// This should not panic
	SafeLogf("test message %s", "value")
	
	// Test with valid logger
	// Note: We can't easily test log output without capturing it,
	// but we can at least verify it doesn't crash
	Log = originalLog
	SafeLogf("test message %s", "value")
}

func TestSafeLog(t *testing.T) {
	// Test with nil logger (should not crash)
	originalLog := Log
	Log = nil
	
	// This should not panic
	SafeLog("test message")
	
	// Test with valid logger
	Log = originalLog
	SafeLog("test message")
}

// Helper function since strings.Contains might not be available in test context
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}