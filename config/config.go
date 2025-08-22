package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RPCEndpoint     string
	CacheTTL        time.Duration
	RateLimitPerMin int
	MongoURI        string
	MongoDBName     string
}

func LoadConfigFromEnv() (Config, error) {
	endpoint := os.Getenv("RPC_ENDPOINT")
	if endpoint == "" {
		return Config{}, fmt.Errorf("RPC_ENDPOINT must be set")
	}

	ttl := 10 * time.Second
	if v := os.Getenv("CACHE_TTL"); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			ttl = parsed
		} else {
			return Config{}, fmt.Errorf("invalid CACHE_TTL: %w", err)
		}
	}

	limit := 10
	if v := os.Getenv("RATE_LIMIT_PER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		} else {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_PER_MIN: %q", v)
		}
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return Config{}, fmt.Errorf("MONGO_URI must be set")
	}

	mongoDB := os.Getenv("MONGO_DBNAME")
	if mongoDB == "" {
		mongoDB = "test" 
	}

	return Config{
		RPCEndpoint:     endpoint,
		CacheTTL:        ttl,
		RateLimitPerMin: limit,
		MongoURI:        mongoURI,
		MongoDBName:     mongoDB,
	}, nil
}
