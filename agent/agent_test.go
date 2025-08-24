package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/superwhys/litegate/auth"
	"github.com/superwhys/litegate/config"
)

func TestServeHTTP_ProxiesRequestWithoutAuth(t *testing.T) {
	// upstream server
	app := gin.Default()
	app.GET("/api/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Host":  c.Request.Host,
			"Path":  c.Request.URL.Path,
			"Query": c.Request.URL.Query(),
		})
	})
	upstream := httptest.NewServer(app)
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodGet, "http://proxy.example.com/__proxyServer/api/hello?x=1", nil)
	rr := httptest.NewRecorder()

	a, err := NewAgent(&config.Upstream{
		UpstreamURL: upstream.URL,
		TargetPath:  "/api/hello",
		Timeout:     0,
	})
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}
	a.ServeHTTP(rr, req)

	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	var respBody map[string]any
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	assert.Equal(t, "proxy.example.com", respBody["Host"])
	assert.Equal(t, "/api/hello", respBody["Path"])
	assert.Equal(t, []any{"1"}, respBody["Query"].(map[string]any)["x"])
}

func TestServeHTTP_InjectsAuthData(t *testing.T) {
	app := gin.Default()
	app.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Header": gin.H{
				"X-User": c.GetHeader("X-User"),
			},
			"Query": c.Request.URL.Query(),
		})
	})
	upstream := httptest.NewServer(app)
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

	a, err := NewAgent(&config.Upstream{
		Auth:        authConfig,
		UpstreamURL: upstream.URL,
		TargetPath:  "/api",
		Timeout:     0,
	})
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://proxy.example.com/hello", nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	var respBody map[string]any
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	assert.Equal(t, "alice", respBody["Header"].(map[string]any)["X-User"])
	assert.Equal(t, []any{"42"}, respBody["Query"].(map[string]any)["user_id"])
}

func TestServeHTTP_TimeoutReturnsBadGateway(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow"))
	}))
	defer upstream.Close()

	a, err := NewAgent(&config.Upstream{
		UpstreamURL: upstream.URL,
		TargetPath:  "/",
		Timeout:     50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://proxy.example.com/slow", nil)
	rr := httptest.NewRecorder()
	a.ServeHTTP(rr, req)

	resp := rr.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}
