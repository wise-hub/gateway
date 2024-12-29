package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fibank.bg/fis-gateway-ws/internal/configuration"
	"fibank.bg/fis-gateway-ws/internal/filter"
	"fibank.bg/fis-gateway-ws/internal/routes"
	"fibank.bg/fis-gateway-ws/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize configuration and dependencies
	d, err := configuration.Init()
	if err != nil {
		logrus.Fatalf("Failed to initialize configuration: %v", err)
	}
	defer func() {
		if d.Db != nil {
			d.Db.Close() // Ensure the database pool is closed on shutdown
		}
	}()

	// Set up caching and logging
	util.UserCache = util.NewCache()
	filter.SetupLogger(d.Cfg.LoggerType)

	// Set up routes
	r := chi.NewRouter()
	r.Use(filter.CORSMiddleware)
	r.Use(middleware.RequestLogger(&filter.LogFormatter{}))
	r.Use(middleware.Recoverer)
	routes.SetupRoutes(r, d)

	// Configure the HTTP server
	s := &http.Server{
		Addr:              ":" + d.Cfg.Port,
		Handler:           r,
		ReadTimeout:       d.Cfg.ReadTimeout * time.Second,
		WriteTimeout:      d.Cfg.WriteTimeout * time.Second,
		IdleTimeout:       d.Cfg.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.SetKeepAlivesEnabled(true)

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logrus.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			logrus.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	// Log startup information
	logStartupInfo(d.Cfg.EnvType, d.Cfg.Port)

	// Start the server
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Server failed: %v", err)
	}
}

// Log server startup details
func logStartupInfo(envType, port string) {
	startupMsg := "\n-----------------------------------------------------------\n" +
		"Starting server in %s environment on port %s\n" +
		"-----------------------------------------------------------"
	log.Printf(startupMsg, envType, port)
	logrus.Infof(startupMsg, envType, port)
}
