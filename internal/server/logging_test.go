package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ibinds/internal/config"
)

func TestLoggingIncludesErrorDetailWithoutToken(t *testing.T) {
	var buf bytes.Buffer
	s := New(&config.Config{ProfilesDir: t.TempDir()}, ServerOpts{
		Token:  "secret-token",
		Logger: log.New(&buf, "", 0),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health?token=wrong-token", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	out := buf.String()
	if !strings.Contains(out, "GET /api/health -> 401") {
		t.Fatalf("log missing request line: %q", out)
	}
	if !strings.Contains(out, "unauthorized") {
		t.Fatalf("log missing error detail: %q", out)
	}
	if strings.Contains(out, "secret-token") || strings.Contains(out, "wrong-token") {
		t.Fatalf("log leaked the token: %q", out)
	}
}
