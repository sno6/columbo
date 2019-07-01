package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sno6/columbo/fingerprint"

	"github.com/julienschmidt/httprouter"
)

const (
	defaultAddr         = ":8080"
	defaultReadTimeout  = time.Second * 5
	defaultWriteTimeout = time.Second * 10
)

// Server for columbo.
type Server struct {
	srv *http.Server
}

// An Option changes the default configurations of the server.
type Option func(*Server)

// New creates a server with default settings unless options are provided.
func New(store fingerprint.Store, opts ...Option) (*Server, error) {
	r := httprouter.New()
	registerHandlers(r, store)
	s := &Server{
		srv: &http.Server{
			Addr:         defaultAddr,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			Handler:      r,
		},
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// Start accepting connections from clients.
func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// Host returns the host address of the server.
func (s *Server) Host() string {
	addr := s.srv.Addr
	if addr[0] == ':' {
		return "127.0.0.1" + addr
	}
	return addr
}

// Addr is an option to set the address of the server.
func Addr(a string) Option {
	return func(s *Server) {
		s.srv.Addr = a
	}
}

// ReadTimeout is an option to set how long the server should take reading a request.
func ReadTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.srv.ReadTimeout = t
	}
}

// WriteTimeout is an option to set how long the server should take writing a response.
func WriteTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.srv.WriteTimeout = t
	}
}

func registerHandlers(r *httprouter.Router, store fingerprint.Store) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pub := filepath.Join(wd, "/web/public")

	// TODO: If possible, remove this filesystem read for index.html on every request.
	r.ServeFiles("/public/*filepath", http.Dir(pub))
	r.Handler(http.MethodGet, "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(pub, "/bundle/index.html"))
	}))

	r.Handler(http.MethodGet, "/ws", &WS{Store: store, logger: log.New(os.Stdout, "", log.LstdFlags)})
}
