package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-redis/redis/v9"

	_ "github.com/joho/godotenv/autoload"
)

// NOTE: The underscore before `github.com/joho/godotenv/autoload` autoloads the .env if available

// Create a client for the Redis server based on current deploy environment
var client = getRedisClient()

func main() {
	proxy := getProxy()
	mux := getMux()

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
