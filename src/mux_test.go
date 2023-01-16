package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-redis/redismock/v9"
)

func Test_mux_OPTIONS(t *testing.T) {
	// ARRANGE
	proxy := getProxy()
	mux := getMux(proxy)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()

	// ACT
	mux.ServeHTTP(w, req)

	// ARRANGE
	res := w.Result()
	receivedStatusCode := res.StatusCode
	receivedHeaders := res.Header

	// ASSERT
	if receivedStatusCode != http.StatusNoContent {
		t.Fatalf("Status code for OPTIONS not 204, instead got %d", receivedStatusCode)
	}

	expectedHeaders := http.Header{
		"Access-Control-Allow-Origin":      {"http://localhost:3000"},
		"Access-Control-Allow-Methods":     {"POST"},
		"Access-Control-Allow-Headers":     {"Content-Type"},
		"Access-Control-Allow-Credentials": {"true"},
	}

	if !reflect.DeepEqual(expectedHeaders, receivedHeaders) {
		t.Fatalf("Got incorrect headers. \nExpected %s \nReceived %s", expectedHeaders, receivedHeaders)
	}
}

func Test_mux_Cache(t *testing.T) {
	// ARRANGE
	proxy := getProxy()
	mux := getMux(proxy)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	expectedKey, _ := generateCacheKeyFromRequest(req)

	w := httptest.NewRecorder()

	db, mock := redismock.NewClientMock()

	client = db

	// ACT
	mux.ServeHTTP(w, req)

	mock.ExpectGet(expectedKey).SetVal("sdf")

}
