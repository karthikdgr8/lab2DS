package main

import "os"

var serverIP string
var serverPort string

func main() {
	port := os.Args[1]
	serverIP = os.Getenv("SERVER_IP")
	serverPort = os.Getenv("SERVER_PORT")

	if serverIP == "" {
		serverIP = os.Args[2]
	}
	if serverPort == "" {
		serverPort = os.Args[3]
	}

	startProxy(port)
}
