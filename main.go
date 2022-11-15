package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
)

func main() {
	port := os.Args[1]
	startServer(port)
}

func startServer(port string) {
	fmt.Print("Server starting on port: " + port)
	server, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Print("Error starting server: " + err.Error())
		os.Exit(-1)
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting new client: " + err.Error())
		} else {
			go processClient(client)
		}
	}
}

func processClient(client net.Conn) {
	println("\nNew client accepted, processing request.")
	buf := make([]byte, 0) // do not use.
	tmp := make([]byte, 10048)
	buffer := bytes.NewBuffer(buf)
	readLen, err := client.Read(tmp)
	if err != nil {
		fmt.Println("Error")
	}
	//for err == nil && readLen != -1 {
	buffer.Write(tmp[:readLen])
	//fmt.Print("Read: ")
	//fmt.Println(tmp[:readLen])
	//	break
	//readLen, err = client.Read(tmp)
	//}

	reader := bufio.NewReader(buffer)
	req, err := http.ReadRequest(reader)

	println("URL: " + req.URL.Path)
	fmt.Println("Got request: " + req.Method)
	if req.Method == "GET" { // handle get

	} else if req.Method == "POST" { //handle post

	} else {
		var res = http.Response{Close: true, StatusCode: 501}
		res.Write(client)

		req.Close = true
	}
}
