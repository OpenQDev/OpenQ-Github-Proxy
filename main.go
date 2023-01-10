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
	"strconv"
	"strings"

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

	// Modify response
	b = bytes.Replace(b, []byte("server"), []byte("schmerver"), -1)
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))

	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	resp.Header.Set("Access-Control-Allow-Origin", "http://localhost:3000")
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	resp.Header.Set("Access-Control-Allow-Credential", "true")
	return resp, nil
}

var _ http.RoundTripper = &transport{}

func main() {
	godotenv.Load()

	target, err := url.Parse("https://api.github.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &transport{http.DefaultTransport}

	http.Handle("/", proxy)

	fmt.Println("Listening on port 3005")

	// Start the server using the mux wrapped with CORs package to append necessary headers
	http.ListenAndServe(":3005", nil)
}
