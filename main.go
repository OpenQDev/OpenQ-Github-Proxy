package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:8081

func main() {
	// Create a proxy server
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "https",
		Host:   "api.github.com",
	})

	// Create a handler function for the proxy server
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Authorization", "Bearer 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")

		// Set the URL path to the GraphQL endpoint
		r.URL.Path = "/graphql"

		// Set the Host header to the host of the GraphQL API
		r.Host = "api.github.com"

		// Serve the request through the proxy
		proxy.ServeHTTP(w, r)
	})

	// Start the server
	http.ListenAndServe(":3005", nil)

}
