package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// maskDatabaseURL hides passwords in logs (best-effort).
func maskDatabaseURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	name := u.User.Username()
	u.User = url.UserPassword(name, "****")
	return u.String()
}

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	ITunesBaseURL string
	ITunesMock    bool
	CacheTTL      time.Duration
	Environment   string
}

// Load pulls settings from the environment; loads a local .env when it exists, then applies sane defaults in dev.
func Load() Config {
	_ = godotenv.Load()

	env := get("ENV", "development")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" && env == "development" {
		dbURL = get(
			"DATABASE_URL_DEFAULT",
			"postgres://ccp:ccp@localhost:5432/ccp?sslmode=disable",
		)
		log.Printf(
			"config: DATABASE_URL not set; using development default %s (override with DATABASE_URL or .env)",
			maskDatabaseURL(dbURL),
		)
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required when ENV is not development (copy .env.example to .env)")
	}

	itunesMock := getBool("ITUNES_MOCK", false)
	if !itunesMock {
		itunesMock = getBool("OPEN_LIBRARY_MOCK", false)
	}

	return Config{
		HTTPAddr:      get("HTTP_ADDR", ":8080"),
		DatabaseURL: dbURL,
		ITunesBaseURL: get("ITUNES_BASE_URL", "https://itunes.apple.com"),
		ITunesMock:    itunesMock,
		CacheTTL:      getDuration("CACHE_TTL_SECONDS", 30*time.Second),
		Environment:   env,
	}
}

// get returns os.Getenv(key) when set, otherwise def.
func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// getDuration treats the env value as whole seconds; non-numeric input keeps def.
func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return def
}
