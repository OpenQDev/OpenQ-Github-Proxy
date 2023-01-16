package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func getProxy() *httputil.ReverseProxy {
	target, err := url.Parse("https://api.github.com")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &transport{http.DefaultTransport}

	return proxy
}
