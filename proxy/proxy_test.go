package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/superwhys/litegate/auth"
	"github.com/superwhys/litegate/config"
)

func TestServeHTTP_ProxiesRequestWithoutAuth(t *testing.T) {
	received := make(chan *http.Request, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	a, err := NewAgent(nil, upstream.URL, "/api", 0)
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example/hello?x=1", nil)
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	select {
	case r := <-received:
		t.Logf("received request: %s", r.URL.String())
		assert.Equal(t, "/api/hello", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("x"))
	default:
		t.Fatalf("upstream did not receive request")
	}
}

func TestServeHTTP_InjectsAuthData(t *testing.T) {
	received := make(chan *http.Request, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	authConfig := &config.Auth{
		Type:   "jwt",
		Source: "$header.Authorization",
		Secret: "secret",
		Claims: map[string]string{
			"$header.X-User": "userName",
			"$query.user_id": "userId",
		},
	}

	a, err := NewAgent(authConfig, upstream.URL, "/api", 0)
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example/hello", nil)
	// 这里就不验证 a.Auth() 了，所以这里直接注入 claims 到 context 中
	ctx := context.WithValue(req.Context(), auth.ClaimContextKey("$header.X-User"), "alice")
	ctx = context.WithValue(ctx, auth.ClaimContextKey("$query.user_id"), "42")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	select {
	case r := <-received:
		t.Logf("received request: %s, header: %v, query: %v", r.URL.String(), r.Header, r.URL.Query())
		assert.Equal(t, "alice", r.Header.Get("X-User"))
		assert.Equal(t, "42", r.URL.Query().Get("user_id"))
	default:
		t.Fatalf("upstream did not receive request")
	}
}

func TestServeHTTP_TimeoutReturnsBadGateway(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow"))
	}))
	defer upstream.Close()

	a, err := NewAgent(nil, upstream.URL, "/", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example/slow", nil)
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	resp := rr.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}
