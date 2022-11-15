package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

const SERVER_PORT = "8080"

func main() {
	startServer()
}

func startServer() {
	fmt.Print("Server starting on port: " + SERVER_PORT)
	server, err := net.Listen("tcp", "localhost:"+SERVER_PORT)
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
	buf := make([]byte, 0) // do not use.
	tmp := make([]byte, 1024)
	buffer := bytes.NewBuffer(buf)
	readLen, err := client.Read(tmp)
	for err == nil && readLen != -1 {
		buffer.Write(tmp[:readLen])
		fmt.Print("Read: ")
		fmt.Println(tmp[:readLen])
		break
		//readLen, err = client.Read(tmp)
	}

	reader := bufio.NewReader(buffer)
	req, err := http.ReadRequest(reader)
	fmt.Println("Got request: " + req.Method)
	if req.Method != "GET" {
		var res = http.Response{Close: true, StatusCode: 501}

		i := 1
		mybytes, err := io.ReadAll(res.Body)

		if err != nil {
			fmt.Println("Error reading res body for response")
		}
		client.Write(mybytes)
		req.Close = true
	}
}
