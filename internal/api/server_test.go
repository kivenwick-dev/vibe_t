package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"banner-fingerprint/internal/fingerprint"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	engine, err := fingerprint.Load(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	return NewServer(engine, slog.Default()).Handler()
}

func TestHealth(t *testing.T) {
	response := httptest.NewRecorder()
	testHandler(t).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestFingerprintUnknownDoesNotFailBatch(t *testing.T) {
	body := bytes.NewBufferString(`[{"ip":"10.0.0.1","port":9999,"banner":"opaque"}]`)
	request := httptest.NewRequest(http.MethodPost, "/fingerprint", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	testHandler(t).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	var results []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["protocol"] != "unknown" {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestFingerprintRejectsInvalidJSON(t *testing.T) {
	response := httptest.NewRecorder()
	testHandler(t).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/fingerprint", bytes.NewBufferString("{")))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}
