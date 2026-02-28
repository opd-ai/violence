package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/network"
	"github.com/sirupsen/logrus"
)

// Server configuration flags
var (
	port     = flag.Int("port", 7777, "Server port to listen on")
	logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
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
		"port":      *port,
		"log_level": *logLevel,
	}).Info("Starting VIOLENCE dedicated server")

	// Initialize game world
	world := engine.NewWorld()

	// Create and start game server
	server, err := network.NewGameServer(*port, world)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create game server")
	}

	if err := server.Start(); err != nil {
		logrus.WithError(err).Fatal("Failed to start game server")
	}

	logrus.Info("Server started successfully, waiting for connections...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logrus.Info("Shutdown signal received, stopping server...")

	if err := server.Stop(); err != nil {
		logrus.WithError(err).Error("Error during server shutdown")
	}

	logrus.Info("Server stopped")
}
