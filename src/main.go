package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"authzen/api"
	"authzen/policy"
)

func main() {
	// Parse command line arguments
	var (
		port    = flag.Int("port", 8080, "Server port")
		baseURL = flag.String("base-url", "http://localhost:8080", "Base URL for the server")
		tlsFlag = flag.Bool("tls", false, "Enable TLS")
		cert    = flag.String("cert", "server.crt", "TLS certificate file")
		key     = flag.String("key", "server.key", "TLS key file")
	)
	flag.Parse()

	// Initialize policy store
	store := policy.NewStore()

	// Add sample policies
	store.AddPolicy("user:alice", "document:123", "read", true)
	store.AddPolicy("user:alice", "document:123", "write", true)
	store.AddPolicy("user:bob", "document:123", "read", true)
	store.AddPolicy("user:bob", "document:123", "write", false)
	store.AddPolicy("user:charlie", "document:123", "read", false)

	// Initialize API server
	server := api.NewServer(store, *baseURL)

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server
	go startServer(server, *port, *tlsFlag, *cert, *key)

	// Wait for signal
	sig := <-sigCh
	log.Printf("Received signal: %v", sig)
	log.Println("Shutting down server...")
}

// startServer starts the server
func startServer(server *api.Server, port int, tlsEnabled bool, certFile, keyFile string) {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server: %s", addr)

	// Set up HTTPS server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Router(),
	}

	// If TLS is enabled
	if tlsEnabled {
		// TLS configuration
		httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		// Check if certificate files exist
		_, certErr := os.Stat(certFile)
		_, keyErr := os.Stat(keyFile)

		if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
			log.Printf("Certificate files not found. Disabling HTTPS.")
			tlsEnabled = false
		}
	}

	var err error
	if tlsEnabled {
		log.Printf("Starting HTTPS Authorization API server on port: %d", port)
		err = httpServer.ListenAndServeTLS(certFile, keyFile)
	} else {
		log.Printf("Starting HTTP Authorization API server on port: %d", port)
		err = httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
