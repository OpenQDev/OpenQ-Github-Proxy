package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

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

	cacheKey, err := generateCacheKeyFromRequest(req)
	if err != nil {
		log.Panic(err)
	}

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
