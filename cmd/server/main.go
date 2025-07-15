package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"pdfmerge/internal/handler"
	"pdfmerge/internal/repository"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Define command line flags
	portFlag := flag.String("port", "", "Port to run the server on (overrides PORT env var)")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	if err := godotenv.Load(); err != nil {
		sugar.Warn("No .env file found, using defaults")
	}

	// Port priority: command line flag > environment variable > default
	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		sugar.Fatal("BASE_URL must be set in .env")
	}

	repo := repository.NewPDFRepository()
	mergeHandler := handler.NewMergeHandler(repo, sugar)

	mux := http.NewServeMux()
	mux.Handle("/merge", mergeHandler)

	reportHandler := handler.NewReportHandler(repo, sugar, baseURL)
	mux.Handle("GET /report/{id}", reportHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		sugar.Infof("Server started on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	sugar.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		sugar.Fatalf("Graceful shutdown failed: %v", err)
	}

	sugar.Info("Server shut down cleanly")
}
