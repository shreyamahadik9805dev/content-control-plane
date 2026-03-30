package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shreyafeo/content-control-plane/internal/cache"
	"github.com/shreyafeo/content-control-plane/internal/client/itunes"
	"github.com/shreyafeo/content-control-plane/internal/config"
	"github.com/shreyafeo/content-control-plane/internal/handler"
	"github.com/shreyafeo/content-control-plane/internal/repository"
	"github.com/shreyafeo/content-control-plane/internal/service"
)

func main() {
	// Wire-up lives here so the rest of the tree stays easy to test in isolation.
	cfg := config.Load()
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	repo, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// go-cache uses a background sweep interval; keep it reasonably above the entry TTL.
	cleanup := cfg.CacheTTL * 2
	if cleanup < time.Minute {
		cleanup = time.Minute
	}
	ttlCache := cache.New(cfg.CacheTTL, cleanup)
	itClient := itunes.New(cfg.ITunesBaseURL, cfg.ITunesMock)
	podcasts := service.NewPodcasts(repo, ttlCache, itClient)
	h := handler.New(podcasts)

	r := gin.Default()
	h.Register(r)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
