package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// curl -X POST -H "Content-Type: application/json" -d '{"query": "query { repository(name: \"OpenQ-Frontend\", owner: \"OpenQDev\") { issue(number: 124) { title } } }"}' http://localhost:3005

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a client for the Redis server
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

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

		// Generate a cache key for the request
		// Read the request body
		// defer r.Body.Close()
		// body, _ := ioutil.ReadAll(r.Body)

		// Convert the request body to a JSON string
		// jsonString, err := convertToJSONString(body)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		fmt.Printf("%s", body)

		// Parse the request body as a GraphQL query
		var query string
		err = json.Unmarshal(body, &query)
		if err != nil {
			http.Error(w, "Error parsing request body as GraphQL query", http.StatusBadRequest)
			return
		}

		cacheKey, err := json.Marshal(query)
		if err != nil {
			http.Error(w, "Error marshalling GraphQL query to JSON", http.StatusInternalServerError)
			return
		}

		// give me a random integer and convert it to a string
		key := r.URL.String() + r.Method + string(cacheKey)

		// Check if the response is in the cache
		if val, err := client.Get(r.Context(), key).Result(); err == redis.Nil {
			// Create a ResponseRecorder to capture the response
			recorder := httptest.NewRecorder()

			// Response not in cache, serve the request through the proxy
			proxy.ServeHTTP(recorder, r)

			// Serialize the response to a byte slice
			res, _ := httputil.DumpResponse(recorder.Result(), true)

			// Store the response in the cache
			client.Set(r.Context(), key, res, 0)

			// Write the serialized response to the ResponseWriter
			w.Write(res)
		} else if err != nil {
			// Error occurred while fetching from cache
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			// Response found in cache, serve it to the client
			w.Write([]byte(val))
		}
	})

	// Start the server
	http.ListenAndServe(":3005", nil)
}

func convertToJSONString(body []byte) (string, error) {
	// Unmarshal the request body to a map
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	// Marshal the map to a JSON string
	jsonString, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
