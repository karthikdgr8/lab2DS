package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const goServerIp = "localhost"
const goServerPort = 8080

func callServer(r *http.Request) http.Response {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	resp, err := c.Do(r)
	if err != nil {
		fmt.Printf("Error %s", err)
		return http.Response{StatusCode: http.StatusInternalServerError}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Printf("Body : %s", body)
	return *resp
}
