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
	"fibank.bg/fis-gateway-ws/internal/middleware_custom"
	"fibank.bg/fis-gateway-ws/internal/routes"
	"fibank.bg/fis-gateway-ws/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	d, err := configuration.Init()
	if err != nil {
		panic(err)
	}

	// Initialize the UserCache
	util.UserCache = util.NewCache()

	r := chi.NewRouter()

	// Applying middlewares
	r.Use(middleware_custom.CORSMiddleware)
	r.Use(middleware_custom.LogMiddleware(d))
	r.Use(middleware.Recoverer)

	// Setup routes
	routes.SetupRoutes(r, d)

	s := &http.Server{
		Addr:           ":" + d.Cfg.Port,
		Handler:        h2c.NewHandler(r, &http2.Server{}),
		ReadTimeout:    d.Cfg.ReadTimeout * time.Second,
		WriteTimeout:   d.Cfg.WriteTimeout * time.Second,
		IdleTimeout:    d.Cfg.IdleTimeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		d.AccessLogger.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			d.ErrorLogger.Fatal("Server forced to shutdown: ", err)
		}
	}()

	d.AccessLogger.Println("-----------------------------------------------------------")
	d.AccessLogger.Printf("Starting server in %s environment on port %s\n", d.Cfg.EnvType, d.Cfg.Port)
	d.AccessLogger.Println("-----------------------------------------------------------")

	log.Printf("Starting server in %s environment on port %s\n", d.Cfg.EnvType, d.Cfg.Port)

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
