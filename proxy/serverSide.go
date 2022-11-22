package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const goServerIp = "localhost"
const goServerPort = 8080

func callServer() {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req := http.Request{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Printf("Body : %s", body)
}
