package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:3005

func main() {
	godotenv.Load()

	proxy := &httputil.ReverseProxy{
		// A Director is used to modify the request before sending to target server
		Director: func(r *http.Request) {
			const OAUTH_TOKEN_COOKIE_NAME string = "github_oauth_token_unsigned"

			// Check for the "github_oauth_token_unsigned" cookie
			if cookie, err := r.Cookie(OAUTH_TOKEN_COOKIE_NAME); err == nil {
				// Add the cookie value as the Authorization header if present
				r.Header.Set("Authorization", "Bearer "+cookie.Value)
			} else {
				// Add a default Authorization header if not present
				commaDelimitedPATs := os.Getenv("PATS")
				pats := strings.Split(commaDelimitedPATs, ",")
				index := rand.Intn(len(pats))
				randomPat := pats[index]

				r.Header.Set("Authorization", "Bearer "+randomPat)
			}

			// Set the URL path to the GraphQL endpoint
			r.URL.Scheme = "https"

			r.URL.Path = "/graphql"

			// Set the Host Header AND the URL Host to the GraphQL API endpoint (https://github.com/golang/go/issues/28168)
			r.Host = "api.github.com"
			r.URL.Host = "api.github.com"
		},
		// ModifyResponse is used to modify the response from the target server before sending it back to the client
		ModifyResponse: func(r *http.Response) error {
			// store response in cache here
			return nil
		},
	}

	mux := http.NewServeMux()

	// Create a Handler function on the mux to check cache before passing request to Proxy
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		const CacheHit bool = false

		if CacheHit {

		} else {
			// Response not in cache, serve the request through the proxy
			proxy.ServeHTTP(w, r)
		}
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
	})

	handler := c.Handler(mux)

	fmt.Println("Listening on port 3005")

	// Start the server using the mux wrapped with CORs package to append necessary headers
	http.ListenAndServe(":3005", handler)
}
