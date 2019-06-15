package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomiddleware/recover"
	"github.com/oschwald/geoip2-golang"
)

type GeoDatabase struct {
	db Database
}

func NewGeoDatabase(path string) (*GeoDatabase, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &GeoDatabase{db}, nil
}

func (g GeoDatabase) City(ip net.IP) (*geoip2.City, error) {
	return g.db.City(ip)
}

func main() {
	// Create a config from env variables with a prefix
	cfg, err := newCfg("")
	if err != nil {
		log.Fatalf("main: Error loading config: %s", err.Error())
	}

	// Create a logger
	log := NewLogger(os.Stdout, cfg.Log.Level, cfg.Log.Dev)

	// Open the database
	db, err := NewGeoDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}

	// create middlewares for our api server
	// by createing a list of middlewares to enable on http server
	mw := []func(http.Handler) http.Handler{
		contentType("application/json"),
		loggerMiddleware(log),
		recover.Handler,
	}

	// create the http.Server
	api := http.Server{
		Addr:           cfg.Web.Host + ":" + cfg.Web.Port,
		Handler:        Chain(NewGeoServer(log, *db), mw...),
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Listening channel for errors
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Debugf("Starting api Listening %s:%s", cfg.Web.Host, cfg.Web.Port)
		serverErrors <- api.ListenAndServe()
	}()

	// ========================================
	// Shutdown
	//
	// Listen for os signals
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// ========================================
	// Stop API Service
	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-osSignals:
		log.Info("Start shutdown...")

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := api.Shutdown(ctx); err != nil {
			log.Infof("Graceful shutdown did not complete in %v: %v", cfg.Web.ShutdownTimeout, err)
			if err := api.Close(); err != nil {
				log.Fatalf("Could not stop http server: %v", err)
			}
		}
	}
}
