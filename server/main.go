package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"zenith/server/handlers"
)

func resolvePort(args []string) (string, error) {
	if len(args) > 2 {
		return "", fmt.Errorf("usage: zenith-server [PORT]")
	}
	if len(args) < 2 {
		return ":8080", nil
	}
	port, err := strconv.Atoi(args[1])
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid port number: %s", args[1])
	}
	return ":" + strconv.Itoa(port), nil
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
		MaxHeaderBytes: 64 << 10,
	}
}

// serve owns the listener and waits for in-flight requests during shutdown.
func serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			_ = server.Close()
		}
		serveErr := <-done
		if err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}

func main() {
	port, err := resolvePort(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	// Keep the original all-interface default; allow binding to loopback or a
	// private interface without changing the existing positional-port CLI.
	addr := net.JoinHostPort(os.Getenv("ZENITH_HOST"), port[1:])
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	// Windows console Ctrl+C/Ctrl+Break are delivered as os.Interrupt.
	// On Linux/macOS SIGTERM also requests a graceful drain.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := newHTTPServer(addr, handlers.NewHandler().Routes())
	log.Printf("Listening on %s", listener.Addr())
	if err := serve(ctx, server, listener, 10*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Print("Shutdown complete")
}
