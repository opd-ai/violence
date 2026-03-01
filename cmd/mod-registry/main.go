// Package main provides a standalone mod registry HTTP server.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/opd-ai/violence/pkg/mod/registry"
	"github.com/sirupsen/logrus"
)

const version = "6.1.0"

var (
	addr        = flag.String("addr", ":8081", "HTTP server address")
	dbPath      = flag.String("db", "mod-registry.db", "SQLite database path")
	storagePath = flag.String("storage", "mod-storage", "Mod storage directory")
	logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	maxModSize  = flag.Int64("max-mod-size", 10*1024*1024, "Maximum mod size in bytes (default 10MB)")
)

func main() {
	flag.Parse()

	// Configure logging
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Invalid log level")
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.WithFields(logrus.Fields{
		"version": version,
		"addr":    *addr,
		"db":      *dbPath,
		"storage": *storagePath,
	}).Info("Starting mod registry server")

	// Open database
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open database")
	}
	defer db.Close()

	// Create registry
	reg, err := registry.NewRegistry(db, *storagePath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize registry")
	}
	defer reg.Close()

	// Configure max mod size
	reg.SetMaxModSize(*maxModSize)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", reg.HandleUpload)
	mux.HandleFunc("/search", reg.HandleSearch)
	mux.HandleFunc("/download/", reg.HandleDownload)
	mux.HandleFunc("/health", handleHealth)

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	// Start listening
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	// Start server in background
	go func() {
		logrus.WithField("addr", listener.Addr().String()).Info("HTTP server listening")
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("HTTP server error")
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logrus.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Server shutdown error")
	}

	logrus.Info("Server stopped")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, version)
}
