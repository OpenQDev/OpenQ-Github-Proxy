package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type transport struct {
	http.RoundTripper
}

func prepareRequestForRedirect(req *http.Request) {
	// Set the URL path to the Github GraphQL endpoint
	req.URL.Scheme = "https"
	req.URL.Path = "/graphql"

	// Set the Host Header AND the URL Host to the GraphQL API endpoint (https://github.com/golang/go/issues/28168)
	req.Host = "api.github.com"
	req.URL.Host = "api.github.com"
}

type RequestBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func extractIds(req *http.Request) ([]string, error) {
	ids := []string{}

	var body RequestBody
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		panic(err)
	}

	cool := body.Variables["bountyId"].(string)
	fmt.Println("sdfsdfsdfsdfsd", cool)

	return ids, nil
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	setAuthorizationHeader(req)
	prepareRequestForRedirect(req)

	cacheKey, err := generateCacheKeyFromRequest(req)
	extractIds(req)

	if err != nil {
		log.Panic(err)
	}

	// Make request
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	cacheResponse(cacheKey, req, resp)

	// Append CORS headers
	resp.Header.Set("Access-Control-Allow-Origin", os.Getenv("ORIGIN"))
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	resp.Header.Set("Access-Control-Allow-Credentials", "true")
	resp.Header.Set("Access-Control-Allow-Methods", "POST")

	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Content-Encoding", "gzip")

	return resp, nil
}

func cacheResponse(cacheKey string, req *http.Request, resp *http.Response) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rerr := client.Set(req.Context(), cacheKey, b, 1*time.Hour).Err()
	if rerr != nil {
		panic(err)
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	// Reattach body to resp
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body

	return nil
}
