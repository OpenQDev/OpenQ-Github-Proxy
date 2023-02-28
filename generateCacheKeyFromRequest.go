package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
)

func generateCacheKeyFromRequest(req *http.Request) (string, error) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(append(reqBody, []byte(req.URL.Path)...))
	cacheHex := h.Sum(nil)
	cacheKey := hex.EncodeToString(cacheHex)

	err = req.Body.Close()
	if err != nil {
		return "", err
	}

	newBody := ioutil.NopCloser(bytes.NewReader(reqBody))
	req.Body = newBody

	return cacheKey, nil
}
