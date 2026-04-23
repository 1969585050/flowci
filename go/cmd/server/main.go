package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/flowci/flowci/internal/api"
	"github.com/flowci/flowci/internal/builder"
	"github.com/flowci/flowci/internal/config"
	"github.com/flowci/flowci/internal/deployer"
	"github.com/flowci/flowci/pkg/docker"
)

const (
	Version = "0.1.0"
)

func getServerAddr() string {
	if addr := os.Getenv("FLOWCI_API_ADDR"); addr != "" {
		return addr
	}
	return "localhost:3847"
}

func main() {
	addr := getServerAddr()
	fmt.Printf("FlowCI API Server v%s\n", Version)
	fmt.Println("=======================")
	fmt.Printf("Server address: %s\n", addr)

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Config load error (using defaults): %v", err)
		cfg = config.Default()
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Docker connection failed: %v", err)
	}
	defer dockerClient.Close()

	log.Println("Docker connected successfully")

	b := builder.NewBuilder(dockerClient)
	d := deployer.NewDeployer(dockerClient)
	cm := config.NewManager(cfg)

	server := api.NewServer(dockerClient, b, d, cm, Version)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:     server.Router(),
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  60,
	}

	go func() {
		log.Printf("API server starting on http://%s", addr)
		log.Printf("Health check: http://%s/health", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
