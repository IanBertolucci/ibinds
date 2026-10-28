package server

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"
)

const maxLoggedBody = 4096

// loggingResponseWriter captures the status and a bounded error body so the
// log contains the same detail the handler returned to the client.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if remaining := maxLoggedBody - w.body.Len(); remaining > 0 {
		if len(b) > remaining {
			w.body.Write(b[:remaining])
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

// withLogging logs one line per request. The query string is deliberately
// omitted because it contains the session token.
func withLogging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		elapsed := time.Since(start).Round(time.Millisecond)
		if status >= 400 {
			logger.Printf("%s %s -> %d in %s: %s", r.Method, r.URL.Path, status, elapsed, strings.TrimSpace(rec.body.String()))
			return
		}
		logger.Printf("%s %s -> %d in %s", r.Method, r.URL.Path, status, elapsed)
	})
}
