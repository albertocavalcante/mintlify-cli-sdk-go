package mintlify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchClient_Search(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("query") != "bazel remote execution" {
			t.Errorf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}

		resp := SearchResult{
			Total: 2,
			Results: []SearchHit{
				{Title: "Remote Execution", Path: "/remote-execution", Snippet: "Configure remote execution...", Score: 0.95},
				{Title: "Remote Cache", Path: "/remote-cache", Snippet: "Set up remote caching...", Score: 0.80},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	client := NewSearchClient(srv.URL, "test-key")
	result, err := client.Search(context.Background(), SearchOptions{
		Query: "bazel remote execution",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("total: got %d, want 2", result.Total)
	}
	if len(result.Results) != 2 {
		t.Fatalf("results: got %d, want 2", len(result.Results))
	}
	if result.Results[0].Title != "Remote Execution" {
		t.Errorf("first result title: got %q, want %q", result.Results[0].Title, "Remote Execution")
	}
}

func TestSearchClient_NoConfig(t *testing.T) {
	client := NewSearchClient("", "")
	_, err := client.Search(context.Background(), SearchOptions{Query: "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
