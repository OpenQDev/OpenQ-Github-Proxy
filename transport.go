package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
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

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func mapIdsToCacheKeys(req *http.Request, cacheKey string) {
	// Get variables off of the request body (GraphQL query)
	var bodyType RequestBody

	// Save body to re-append to req later
	body, _ := ioutil.ReadAll(req.Body)

	// Decode body to JSON
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&bodyType)
	if err != nil {
		panic(err)
	}

	// Variable will either be called id or ids

	// SINGULAR ID
	nullableId := bodyType.Variables["id"]

	if !isNil(nullableId) {
		id := bodyType.Variables["id"].(string)
		client.LPush(req.Context(), id, []string{cacheKey})
	}

	// PLURAL IDs
	nullableIds := bodyType.Variables["ids"]

	if !isNil(nullableIds) {
		ids := bodyType.Variables["ids"].([]interface{})
		for _, v := range ids {
			client.LPush(req.Context(), v.(string), cacheKey)
		}
	}

	// Re-append request body to *http.Request pointer for later modification
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	setAuthorizationHeader(req)
	prepareRequestForRedirect(req)

	cacheKey, err := generateCacheKeyFromRequest(req)
	mapIdsToCacheKeys(req, cacheKey)

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
