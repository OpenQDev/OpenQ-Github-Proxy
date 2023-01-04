package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:3005

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a proxy server here
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "https",
		Host:   "api.github.com",
	})

	// Create a handler function for the proxy server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check for the "github_oauth_token_unsigned" cookie
		if cookie, err := r.Cookie("github_oauth_token_unsigned"); err == nil {
			// Add the cookie value as the Authorization header
			r.Header.Set("Authorization", "Bearer "+cookie.Value)
		} else {
			// Add a default Authorization header
			bearerToken := os.Getenv("BEARER_TOKEN")
			r.Header.Set("Authorization", "Bearer "+bearerToken)
		}

		// Set the URL path to the GraphQL endpoint
		r.URL.Path = "/graphql"

		// Set the Host header to the host of the GraphQL API
		r.Host = "api.github.com"

		// Response not in cache, serve the request through the proxy
		proxy.ServeHTTP(w, r)
	})

	// Start the server
	http.ListenAndServe(":8080", nil)
}
