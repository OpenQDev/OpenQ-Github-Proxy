package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"

	_ "github.com/joho/godotenv/autoload"
)

// NOTE: The underscore before `github.com/joho/godotenv/autoload` autoloads the .env if available

type transport struct {
	http.RoundTripper
}

// Set the default HTTP RoundTripper that will be used by the DefaultServerMux to our custom transport implementation
var _ http.RoundTripper = &transport{}

// Create a client for the Redis server
var client = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	Username: os.Getenv("REDIS_USERNAME"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
	TLSConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
	},
})

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	const OAUTH_TOKEN_COOKIE_NAME string = "github_oauth_token_unsigned"

	// Check for the "github_oauth_token_unsigned" cookie
	if cookie, err := req.Cookie(OAUTH_TOKEN_COOKIE_NAME); err == nil {
		// Add the cookie value as the Authorization header if present
		req.Header.Set("Authorization", "Bearer "+cookie.Value)
	} else {
		// Add a default Authorization header if not present
		commaDelimitedPATs := os.Getenv("PATS")
		pats := strings.Split(commaDelimitedPATs, ",")
		index := rand.Intn(len(pats))
		randomPat := pats[index]

		req.Header.Set("Authorization", "Bearer "+randomPat)
	}

	// Set the URL path to the GraphQL endpoint
	req.URL.Scheme = "https"

	req.URL.Path = "/graphql"

	// Set the Host Header AND the URL Host to the GraphQL API endpoint (https://github.com/golang/go/issues/28168)
	req.Host = "api.github.com"
	req.URL.Host = "api.github.com"

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write([]byte(string(reqBody)))
	cacheHex := h.Sum(nil)
	cacheKey := hex.EncodeToString(cacheHex)

	err = req.Body.Close()
	if err != nil {
		return nil, err
	}

	newBody := ioutil.NopCloser(bytes.NewReader(reqBody))
	req.Body = newBody

	// Make request
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rerr := client.Set(req.Context(), cacheKey, b, 1*time.Hour).Err()
	if rerr != nil {
		panic(err)
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Cache body with request hash as cache key

	// Reattach body to resp
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body

	// Append CORS headers
	resp.Header.Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	resp.Header.Set("Access-Control-Allow-Credentials", "true")
	resp.Header.Set("Access-Control-Allow-Methods", "POST")

	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Content-Encoding", "gzip")

	return resp, nil
}

func main() {
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

	// Use proxy for all calls on DefaultServerMux
	http.Handle("/", mux)

	// Start the server using the DefaultServerMux
	fmt.Println("Listening on port 3005")
	http.ListenAndServe(":3005", nil)
}
