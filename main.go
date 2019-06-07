package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

func main() {
	// create new logger
	log := logrus.New()

	// Create a config from env variables with a prefix
	cfg, err := newCfg("")
	if err != nil {
		log.Fatalf("main: Error loading config: %s", err.Error())
	}

	// Configure logger
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
		log.Error(err.Error())
	}
	if cfg.Log.Debug {
		level = logrus.DebugLevel
	}
	log.SetLevel(level)

	// Adjust logging format
	log.SetFormatter(&logrus.JSONFormatter{})
	if cfg.Log.Dev {
		log.SetFormatter(&logrus.TextFormatter{})
	}

	// Open database file
	db, err := geoip2.Open(cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}

	// create middlewares for our api server
	// by createing a list of middlewares to enable on http server
	mw := []func(http.Handler) http.Handler{
		render.SetContentType(render.ContentTypeJSON),
		loggerMiddleware(log),
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	}

	// create the http.Server
	api := http.Server{
		Addr:           cfg.Web.Host + ":" + cfg.Web.Port,
		Handler:        newServer(log, db, mw...),
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
