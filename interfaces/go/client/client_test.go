package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := New(baseURL)

	if client == nil {
		t.Fatal("New returned nil")
	}
	if client.BaseURL != baseURL {
		t.Errorf("BaseURL = %s, want %s", client.BaseURL, baseURL)
	}
	if client.HTTP == nil {
		t.Error("HTTP client should not be nil")
	}
	if client.HTTP != http.DefaultClient {
		t.Error("HTTP should be default client")
	}
}

func TestClient_ListSessions(t *testing.T) {
	// Create a test server that returns mock session data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path and query params
		if r.URL.Path != "/api/sessions" {
			t.Errorf("path = %s, want /api/sessions", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %s, want 10", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "5" {
			t.Errorf("offset = %s, want 5", r.URL.Query().Get("offset"))
		}

		// Return mock data
		resp := struct {
			Items []Session `json:"items"`
			Total int       `json:"total"`
		}{
			Items: []Session{
				{ID: "sess1", Target: "ws://example.com"},
				{ID: "sess2", Target: "ws://test.com"},
			},
			Total: 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL)
	sessions, total, err := client.ListSessions(10, 5)

	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].ID != "sess1" {
		t.Errorf("sessions[0].ID = %s, want sess1", sessions[0].ID)
	}
	if sessions[1].Target != "ws://test.com" {
		t.Errorf("sessions[1].Target = %s, want ws://test.com", sessions[1].Target)
	}
}

func TestClient_ListSessions_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Items []Session `json:"items"`
			Total int       `json:"total"`
		}{
			Items: []Session{},
			Total: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL)
	sessions, total, err := client.ListSessions(10, 0)

	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

func TestClient_ListSessions_NetworkError(t *testing.T) {
	client := New("http://localhost:1") // invalid port
	_, _, err := client.ListSessions(10, 0)

	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestClient_ListSessions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := New(server.URL)
	_, _, err := client.ListSessions(10, 0)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestClient_ListSessions_ZeroLimitOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "0" {
			t.Errorf("limit = %s, want 0", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "0" {
			t.Errorf("offset = %s, want 0", r.URL.Query().Get("offset"))
		}

		resp := struct {
			Items []Session `json:"items"`
			Total int       `json:"total"`
		}{
			Items: []Session{},
			Total: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := New(server.URL)
	_, _, err := client.ListSessions(0, 0)

	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
}

func TestClient_CustomHTTPClient(t *testing.T) {
	client := New("http://localhost:8080")
	customClient := &http.Client{}
	client.HTTP = customClient

	if client.HTTP != customClient {
		t.Error("HTTP client should be customizable")
	}
}
