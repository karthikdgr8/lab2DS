package main

import "os"

var serverIP string
var serverPort string

func main() {
	port := os.Args[1]
	serverIP = os.Args[2]
	serverPort = os.Args[3]

	startProxy(port)
}
