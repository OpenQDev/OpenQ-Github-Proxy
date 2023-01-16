package main

import "net/http"

func getMux() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}
