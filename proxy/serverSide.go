package main

import (
	"net/http"
	"net/url"
	"time"
)

/*
*
The server side of the proxy. This function is called by the client-thread, and return the response given by the
actual server.
*/
func callServer(r *http.Request) http.Response {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	path := r.URL.Path
	reqURL := url.URL{Scheme: "http", Host: serverIP + ":" + serverPort, Path: path}
	newReq := http.Request{URL: &reqURL}

	resp, err := c.Do(&newReq)
	if err != nil {
		return http.Response{
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Close:      true,
			StatusCode: http.StatusInternalServerError,
		}
	}

	return *resp
}
