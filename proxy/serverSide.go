package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

func callServer(r *http.Request) http.Response {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	path := r.URL.Path
	reqURL := url.URL{Scheme: "http", Host: serverIP + ":" + serverPort, Path: path}
	newReq := http.Request{URL: &reqURL}

	resp, err := c.Do(&newReq)
	if err != nil {
		log.Println(err)
		return http.Response{StatusCode: http.StatusInternalServerError}
	}

	return *resp
}
