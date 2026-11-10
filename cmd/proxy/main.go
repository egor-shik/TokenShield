package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egor-shik/tokenshield/internal/config"
	"github.com/egor-shik/tokenshield/internal/proxy"
	"github.com/egor-shik/tokenshield/internal/cache"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	cachePath := os.Getenv("CACHE_PATH")
	if cachePath == "" {
		cachePath = "cache.db"
	}

	c, err := cache.NewCache(cachePath)
	if err != nil {
		log.Fatalf("cache init: %v", err)
	}
	defer c.Close()

	p, err := proxy.NewProxy(cfg.ApiURL, c)
	if err != nil {
		log.Fatalf("proxy: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           p,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("proxy listening on %s -> forwarding to %s", srv.Addr, cfg.ApiURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Fatalf("server: %v", err)
	case sig := <-stop:
		log.Printf("received %s, shutting down...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
