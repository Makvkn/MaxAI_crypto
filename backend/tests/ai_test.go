//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maxaicrypto/backend/tests/testsupport"
)

func TestGetAIUsage(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, _ := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get usage status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var usage struct {
		Used      int    `json:"used"`
		Limit     int    `json:"limit"`
		Remaining int    `json:"remaining"`
		Plan      string `json:"plan"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &usage); err != nil {
		t.Fatalf("decode usage: %v", err)
	}
	if usage.Limit != 10 {
		t.Fatalf("limit = %d, want 10", usage.Limit)
	}
	if usage.Used != 0 {
		t.Fatalf("used = %d, want 0", usage.Used)
	}
	if usage.Remaining != 10 {
		t.Fatalf("remaining = %d, want 10", usage.Remaining)
	}
	if usage.Plan != "FREE" {
		t.Fatalf("plan = %q, want FREE", usage.Plan)
	}
}

func TestAIConversationLifecycle(t *testing.T) {
	pool := testsupport.Postgres(t)
	router, _, deps := newWalletsTestStack(t, pool)
	accessToken := guestAccessToken(t, router)
	walletID := createSyncedWallet(t, router, accessToken, deps)

	body, _ := json.Marshal(map[string]any{
		"wallet_id": walletID.String(),
		"title":     "ETH review",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create conversation status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var conversation struct {
		ID       string `json:"id"`
		WalletID string `json:"wallet_id"`
		Title    string `json:"title"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &conversation); err != nil {
		t.Fatalf("decode conversation: %v", err)
	}
	if conversation.WalletID != walletID.String() {
		t.Fatalf("wallet_id = %q", conversation.WalletID)
	}
	if conversation.Title != "ETH review" {
		t.Fatalf("title = %q", conversation.Title)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/"+conversation.ID, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get conversation status = %d, body = %s", rec.Code, rec.Body.String())
	}

	messageBody, _ := json.Marshal(map[string]string{
		"content": "Summarize my portfolio",
	})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/ai/conversations/"+conversation.ID+"/messages", bytes.NewReader(messageBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("send message status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", rec.Header().Get("Content-Type"))
	}

	streamBody, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read stream: %v", err)
	}
	if !strings.Contains(string(streamBody), `"type":"completed"`) && !strings.Contains(string(streamBody), `"type": "completed"`) {
		t.Fatalf("stream missing completed event: %s", streamBody)
	}
	if !strings.Contains(string(streamBody), "tool_started") {
		t.Fatalf("stream missing tool_started event: %s", streamBody)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/conversations/"+conversation.ID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list messages status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var messages struct {
		Items []struct {
			Role     string `json:"role"`
			Status   string `json:"status"`
			Response *struct {
				Intent string `json:"intent"`
				Answer string `json:"answer"`
			} `json:"response"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &messages); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	if len(messages.Items) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(messages.Items))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get usage status = %d", rec.Code)
	}
	var usage struct {
		Used int `json:"used"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &usage)
	if usage.Used != 1 {
		t.Fatalf("used = %d, want 1", usage.Used)
	}
}
