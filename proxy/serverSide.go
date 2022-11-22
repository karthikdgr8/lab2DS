package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const goServerIp = "localhost"
const goServerPort = "8080"

func callServer(r *http.Request) http.Response {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	path := r.URL.Path
	reqURL := url.URL{Scheme: "http", Host: "localhost:8080", Path: path}
	newReq := http.Request{URL: &reqURL}
	println("Sending proxied req")
	resp, err := c.Do(&newReq)
	if err != nil {
		fmt.Printf("Error %s", err)
		return http.Response{StatusCode: http.StatusInternalServerError}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Printf("Body : %s", body)
	return *resp
}
