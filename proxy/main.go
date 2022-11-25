package main

import "os"

var serverIP string
var serverPort string

/*
*
Function to start the proxy
*/
func main() {
	port := os.Args[1]
	serverIP = os.Getenv("SERVER_IP")
	serverPort = os.Getenv("SERVER_PORT")

	/*
		If environment variables above are not set, it means we are running locally so we need to expect command line arguments
	*/

	if serverIP == "" {
		serverIP = os.Args[2]
	}
	if serverPort == "" {
		serverPort = os.Args[3]
	}

	startProxy(port)
}
