package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// NOTE: The underscore before `github.com/joho/godotenv/autoload` autoloads the .env if available

// Create a client for the Redis server based on current deploy environment
var client = getRedisClient()

func main() {
	rand.Seed(time.Now().UnixNano())
	proxy := getProxy()
	mux := getMux(proxy)

	// Start the server using the custom mux
	fmt.Println("Listening on port 3005")
	http.ListenAndServe(":3005", mux)
}
