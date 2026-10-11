package bot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSend(t *testing.T) {
	t.Run("send text message without attachments", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/send" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["message"] != "hello" {
				t.Errorf("expected message hello, got %v", body["message"])
			}
			if body["number"] != "+12345" {
				t.Errorf("expected number +12345, got %v", body["number"])
			}
			if _, ok := body["base64_attachments"]; ok {
				t.Errorf("expected no attachments, got %v", body["base64_attachments"])
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		host := strings.TrimPrefix(server.URL, "http://")
		err := send(host, "+12345", "+67890", "hello")
		if err != nil {
			t.Fatalf("send failed: %v", err)
		}
	})

	t.Run("send message with base64 attachment", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["message"] != "stat" {
				t.Errorf("expected message stat, got %v", body["message"])
			}
			att, ok := body["base64_attachments"].([]any)
			if !ok || len(att) != 1 {
				t.Fatalf("expected 1 attachment, got %v", body["base64_attachments"])
			}
			if att[0] != "data:image/png;base64,abc123" {
				t.Errorf("unexpected attachment content: %v", att[0])
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		host := strings.TrimPrefix(server.URL, "http://")
		err := send(host, "+12345", "+67890", "stat", "data:image/png;base64,abc123")
		if err != nil {
			t.Fatalf("send failed: %v", err)
		}
	})
}
