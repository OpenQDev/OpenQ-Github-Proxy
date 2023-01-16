package main

import (
	"crypto/tls"
	"os"

	"github.com/go-redis/redis/v9"
)

func getRedisClient() *redis.Client {
	// Turn on TLS mode if running anywhere except locally
	if os.Getenv("DEPLOY_ENV") == "production" || os.Getenv("DEPLOY_ENV") == "staging" {
		return redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		})
	} else {
		return redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
	}
}
