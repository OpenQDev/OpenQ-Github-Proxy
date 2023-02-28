package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/go-redis/redis/v9"
)

func invalidateEntity(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	// Get list of cache keys including this id
	length := client.LLen(r.Context(), id).Val() - 1
	cacheKeys := client.LRange(r.Context(), id, 0, length)

	err := client.Del(r.Context(), cacheKeys.Val()...).Err()
	if err != nil {
		log.Panic(err)
	}

	addCorsHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func getMux(proxy *httputil.ReverseProxy) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/invalidate_entity", invalidateEntity)

	// Create a Handler function on the mux to check cache before passing request to Proxy
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			addCorsHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		cacheKey, err := generateCacheKeyFromRequest(r)
		if err != nil {
			log.Panic(err)
		}

		// Check if the response is in the cache
		val, err := client.Get(r.Context(), cacheKey).Result()

		if err == redis.Nil {
			// Cache miss
			fmt.Println("Cache miss. Calling Github API")
			proxy.ServeHTTP(w, r)
		} else if err != nil {
			// Error occurred while fetching from cache
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			fmt.Println("Cache hit! Sending cached response.")
			// Response found in cache, serve it to the client
			addCorsHeaders(w)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write([]byte(val))
		}
	})

	return mux
}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
	w.Header().Set("Access-Control-Allow-Headers", "content-type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
}
