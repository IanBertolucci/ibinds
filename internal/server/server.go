package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync/atomic"

	"ibinds/internal/config"
	"ibinds/internal/inputcap"
)

// ServerOpts are injectable dependencies for the HTTP server.
type ServerOpts struct {
	Token         string
	SimCheck      func() (bool, error)
	InputCapturer inputcap.Capturer
}

// Server holds the HTTP routes and the live config pointer.
type Server struct {
	cfg      atomic.Pointer[config.Config]
	token    string
	simCheck func() (bool, error)
	inputCap inputcap.Capturer
	mux      *http.ServeMux
}

// New builds the server. If opts.Token is empty a random token is generated.
func New(cfg *config.Config, opts ServerOpts) *Server {
	token := opts.Token
	if token == "" {
		token = generateToken()
	}
	s := &Server{
		token:    token,
		simCheck: opts.SimCheck,
		inputCap: opts.InputCapturer,
	}
	s.cfg.Store(cfg)
	s.mux = http.NewServeMux()
	s.routes()
	return s
}

// Handler returns the server's root HTTP handler.
func (s *Server) Handler() http.Handler { return s.mux }

// Token returns the effective session token.
func (s *Server) Token() string { return s.token }

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
