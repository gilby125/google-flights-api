package test

import (
	"net/http"
	"net/http/httptest"
)

// SetupHTTPServer creates a test HTTP server and returns the server and client
func SetupHTTPServer(handler http.Handler) (*httptest.Server, *http.Client) {
	ts := httptest.NewServer(handler)
	client := &http.Client{
		Transport: &http.Transport{},
	}
	return ts, client
}
