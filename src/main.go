package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-redis/redis/v9"

	_ "github.com/joho/godotenv/autoload"
)

// NOTE: The underscore before `github.com/joho/godotenv/autoload` autoloads the .env if available

// Create a client for the Redis server
var client = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	Username: os.Getenv("REDIS_USERNAME"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})

func main() {
	// Turn on TLS mode if running anywhere except locally
	if os.Getenv("DEPLOY_ENV") == "production" || os.Getenv("DEPLOY_ENV") == "staging" {
		client.Options().TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	mux := http.NewServeMux()

	target, err := url.Parse("https://api.github.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &transport{http.DefaultTransport}

	// Create a Handler function on the mux to check cache before passing request to Proxy
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			headers := w.Header()
			headers.Add("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
			headers.Add("Access-Control-Allow-Headers", "Content-Type")
			headers.Add("Access-Control-Allow-Methods", "POST")
			headers.Add("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		h := sha256.New()
		h.Write([]byte(string(reqBody)))
		cacheHex := h.Sum(nil)
		cacheKey := hex.EncodeToString(cacheHex)

		err = r.Body.Close()
		if err != nil {
			panic(err)
		}

		newBody := ioutil.NopCloser(bytes.NewReader(reqBody))
		r.Body = newBody

		// Check if the response is in the cache
		val, err := client.Get(r.Context(), cacheKey).Result()

		if err == redis.Nil {
			// Cache miss
			fmt.Println("Cache miss. Calling Github GraphQL API")
			proxy.ServeHTTP(w, r)
		} else if err != nil {
			// Error occurred while fetching from cache
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Println("Cache hit! Sending cached response.")
			// Response found in cache, serve it to the client
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST")

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write([]byte(val))
		}
	})

	// Start the server using the custom mux
	fmt.Println("Listening on port 3005")
	http.ListenAndServe(":3005", mux)
}
