package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayitas/recension/api/internal/auth"
	"github.com/ayitas/recension/api/internal/blobstore"
	"github.com/ayitas/recension/api/internal/config"
	"github.com/ayitas/recension/api/internal/httpapi"
	"github.com/ayitas/recension/api/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()

	var bootstrapKey, bootstrapHash string
	if cfg.Bootstrap {
		bootstrapKey = cfg.BootstrapAPIKey
		bootstrapHash, err = auth.HashPassword(cfg.BootstrapPassword)
		if err != nil {
			log.Fatalf("hash bootstrap password: %v", err)
		}
	}

	pg, err := store.NewPostgres(ctx, cfg.DatabaseURL, bootstrapKey, bootstrapHash)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	var blobs *blobstore.Store
	if cfg.BlobsEnabled() {
		blobs, err = blobstore.New(ctx, blobstore.Config{
			Endpoint:  cfg.S3Endpoint,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
			Bucket:    cfg.S3Bucket,
			UseSSL:    cfg.S3UseSSL,
		})
		if err != nil {
			log.Printf("blobstore unavailable (%v); blob APIs disabled", err)
			blobs = nil
		} else {
			log.Printf("blobstore ready endpoint=%s bucket=%s", cfg.S3Endpoint, cfg.S3Bucket)
		}
	}

	srv := httpapi.New(pg, cfg, blobs)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
	}

	go func() {
		log.Printf("recension api listening on %s env=%s signup=%v bootstrap=%v blobs=%v",
			httpServer.Addr, cfg.Env, cfg.AllowSignup, cfg.Bootstrap, blobs != nil)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
