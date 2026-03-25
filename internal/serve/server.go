package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/lagz0ne/remmd/internal/app"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// Server is the remmd serve runtime: embedded NATS, HTTP reverse proxy to Vite,
// and NATS handlers wired to the real app repos.
type Server struct {
	app        *app.App
	ns         *natsserver.Server
	nc         *nats.Conn
	httpSrv    *http.Server
	viteDir    string
	port       int
	natsWSPort int
}

// Option configures Server construction.
type Option func(*Server)

// WithPort sets the HTTP listen port (default 4312).
func WithPort(p int) Option {
	return func(s *Server) { s.port = p }
}

// WithNATSWSPort sets the NATS WebSocket port (default 0 = auto).
func WithNATSWSPort(p int) Option {
	return func(s *Server) { s.natsWSPort = p }
}

// New creates a Server. Call Start to run it.
func New(application *app.App, viteDir string, opts ...Option) (*Server, error) {
	s := &Server{
		app:     application,
		viteDir: viteDir,
		port:    4312,
	}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// Start runs the server and blocks until ctx is cancelled or a fatal error occurs.
func (s *Server) Start(ctx context.Context) error {
	// 1. Find free ports for NATS WS and Vite (if not explicitly set)
	if s.natsWSPort == 0 {
		s.natsWSPort = freePort()
	}
	vitePort := freePort()

	// 2. Embedded NATS
	storeDir, err := os.MkdirTemp("", "remmd-nats")
	if err != nil {
		return fmt.Errorf("nats tmpdir: %w", err)
	}
	defer os.RemoveAll(storeDir)

	natsOpts := &natsserver.Options{
		ServerName: "remmd",
		DontListen: true,
		JetStream:  true,
		StoreDir:   storeDir,
		Websocket: natsserver.WebsocketOpts{
			Host:  "127.0.0.1",
			Port:  s.natsWSPort,
			NoTLS: true,
		},
	}
	ns, err := natsserver.NewServer(natsOpts)
	if err != nil {
		return fmt.Errorf("nats server: %w", err)
	}
	ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		return fmt.Errorf("nats not ready within 5s")
	}
	s.ns = ns
	slog.Info("nats embedded started", "ws_port", s.natsWSPort)

	// 3. In-process NATS client
	nc, err := nats.Connect("", nats.InProcessServer(ns))
	if err != nil {
		ns.Shutdown()
		return fmt.Errorf("nats connect: %w", err)
	}
	s.nc = nc

	// 4. Register NATS handlers
	registerHandlers(nc, s.app)

	// 5. Spawn Vite dev server (pass NATS WS port via env)
	go func() {
		if err := runViteDev(ctx, s.viteDir, vitePort, s.natsWSPort); err != nil && ctx.Err() == nil {
			slog.Error("vite subprocess failed", "error", err)
		}
	}()

	// 6. HTTP server with reverse proxy to Vite
	viteURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", vitePort))
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(viteURL)
			r.Out.Host = r.In.Host
		},
		FlushInterval: -1,
	}

	mux := http.NewServeMux()
	mux.Handle("/", proxy)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		slog.Info("shutting down serve...")
		s.Shutdown()
	}()

	slog.Info("remmd serve ready", "url", fmt.Sprintf("http://localhost:%d", s.port))
	if err := s.httpSrv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("http: %w", err)
	}
	return nil
}

// Shutdown gracefully tears down all components.
func (s *Server) Shutdown() {
	if s.httpSrv != nil {
		s.httpSrv.Close()
	}
	if s.nc != nil {
		s.nc.Drain()
	}
	if s.ns != nil {
		s.ns.Shutdown()
		s.ns.WaitForShutdown()
	}
}

// freePort asks the OS for an available port.
func freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}
