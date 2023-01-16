package main

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func setAuthorizationHeader(req *http.Request) {
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
}
