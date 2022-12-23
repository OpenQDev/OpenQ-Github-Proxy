package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:8081

func main() {
	originServerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the target URL
		target, err := url.Parse("https://api.github.com")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Append the authorization header
		r.Header.Add("Authorization", "Bearer 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")

		// Create a new reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Update the headers to allow for SSL redirection
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host

		// Proxy the request
		proxy.ServeHTTP(w, r)
	})

	http.ListenAndServe(":8081", originServerHandler)
}
