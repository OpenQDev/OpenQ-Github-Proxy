package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_setAuthorizationHeader(t *testing.T) {
	// ARRANGE
	reqAuthenticated, _ := http.NewRequest("POST", "http://example.com", nil)
	reqUnauthenticated, _ := http.NewRequest("POST", "http://example.com", nil)
	reqAuthenticated.AddCookie(&http.Cookie{Name: "github_oauth_token_unsigned", Value: "foo"})

	// ACT
	setAuthorizationHeader(reqAuthenticated)
	setAuthorizationHeader(reqUnauthenticated)

	// ASSERT
	if reqAuthenticated.Header.Get("Authorization") != "Bearer foo" {
		t.Errorf("Expected Authorization header to have value foo, but had value: %s", reqAuthenticated.Header.Get("Authorization"))
	}

	if !strings.HasPrefix(reqUnauthenticated.Header.Get("Authorization"), "Bearer ghp_") {
		t.Errorf("Expected Authorization header for unauthenticated call to begin with 'Bearer ghp_', but received value: %s", reqUnauthenticated.Header.Get("Authorization"))
	}
}
