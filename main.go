package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/joho/godotenv"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:3005

type transport struct {
	http.RoundTripper
}

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

	// Make request
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	resp.Header = http.Header{}
	resp.Header.Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	resp.Header.Set("Access-Control-Allow-Credentials", "true")
	resp.Header.Set("Access-Control-Allow-Methods", "*")

	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Content-Encoding", "gzip")

	return resp, nil
}

// Set the default HTTP RoundTripper that will be used by the DefaultServerMux to our custom transport implementation
var _ http.RoundTripper = &transport{}

func main() {
	godotenv.Load()

	mux := http.NewServeMux()

	target, err := url.Parse("https://api.github.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &transport{http.DefaultTransport}

	// Create a client for the Redis server
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	fmt.Println(client)

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

		const key = "fsdfoo"

		// Check if the response is in the cache
		val, err := client.Get(r.Context(), key).Result()

		if err == redis.Nil {
			// Cache miss
			proxy.ServeHTTP(w, r)
		} else if err != nil {
			// Error occurred while fetching from cache
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			// Response found in cache, serve it to the client
			w.Write([]byte(val))
		}
	})

	// Use proxy for all calls on DefaultServerMux
	http.Handle("/", mux)

	// Start the server using the DefaultServerMux
	fmt.Println("Listening on port 3005")
	http.ListenAndServe(":3005", nil)
}
