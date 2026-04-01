package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shreyafeo/content-control-plane/internal/cache"
	"github.com/shreyafeo/content-control-plane/internal/client/itunes"
	"github.com/shreyafeo/content-control-plane/internal/config"
	"github.com/shreyafeo/content-control-plane/internal/handler"
	"github.com/shreyafeo/content-control-plane/internal/repository"
	"github.com/shreyafeo/content-control-plane/internal/service"
)

// main is the only composition root: env, Postgres, cache + iTunes, then Gin.
func main() {
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
	// Browsers treat the Vite dev server (another port) as a separate origin; only open this up in development.
	if cfg.Environment == "development" {
		r.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
		})
	}
	h.Register(r)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
