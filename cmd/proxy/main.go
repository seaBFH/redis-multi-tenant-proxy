package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/seabfh/redis-multi-tenant-proxy/internal/proxy"

	"github.com/seabfh/redis-multi-tenant-proxy/internal/config"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadFromFile(*configFile)
	if err != nil {
		// If file doesn't exist, try environment variables
		cfg, err = config.LoadFromEnv()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Create and start proxy
	redisProxy, err := proxy.NewProxy(cfg)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	// Start the proxy in a goroutine
	go func() {
		if err := redisProxy.Start(); err != nil {
			log.Fatalf("Proxy error: %v", err)
		}
	}()

	log.Printf("Redis multi-tenant proxy started on %s", cfg.ListenAddr)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down proxy...")
	redisProxy.Shutdown()
	log.Println("Proxy shutdown complete")
}
