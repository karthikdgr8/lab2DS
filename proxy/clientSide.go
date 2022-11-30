package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"lab1DS/sem"
	"log"
	"net"
	"net/http"
)

type StdResponse struct {
	Code    int
	Error   bool
	Message string
}

/**
Our implementation of a proxy server. This file contains the code for handling incoming connections
to the proxy. Once a connection is established, the request is read, and forwarded to the "serverSide"
Which then forwards the request to the actual webserver.
*/

func startProxy(port string) {

	/*
		Different bind addresses for local and docker instances
	*/
	var bindAddr string
	if serverIP == "" && serverPort == "" {
		bindAddr = "localhost"
	} else {
		bindAddr = "0.0.0.0"
	}

	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", bindAddr+":"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	weightedSem := sem.NewWeighted(10) // Semaphore to ensure no more than 10 simultaneous threads.
	for {
		semErr := weightedSem.Acquire(context.Background(), 1) // Grab one per client
		client, err := server.Accept()
		if err != nil || semErr != nil {
			log.Println("Error accepting new client: " + err.Error())
			sendResponse(http.StatusInternalServerError, true, err.Error(), client) // In case error.
			return
		} else {
			go processClient(client, weightedSem)
		}
	}
}

/*
*
This function handles each new client request, and is run on a separate go-routine.
*/
func processClient(client net.Conn, weighted *sem.Weighted) {
	log.Println("New client accepted,", client.RemoteAddr())
	defer weighted.Release(1)
	defer func(client net.Conn) {
		err := client.Close()
		if err != nil {
			sendResponse(http.StatusInternalServerError, false, "Error Closing conn", client)
		}
	}(client)

	reader := bufio.NewReader(client)
	req, err := http.ReadRequest(reader) // Parse the http-request
	if err != nil {
		sendResponse(http.StatusInternalServerError, true, err.Error(), client)
		return
	}

	log.Println("URL path: " + req.URL.Path)
	log.Println("Request method: " + req.Method)

	if req.Method == "GET" { // handle get

		resp := callServer(req)

		err := resp.Write(client) // Send response to client.
		if err != nil {
			log.Println("Write Err", err)
		}
	} else { //Nothing other than get should be proxied.
		sendResponse(http.StatusNotImplemented, true, "Only GET is implemented", client)
	}

}

// Function to send response to client.
func sendResponse(code int, error bool, message string, client net.Conn) {
	jsonifiedStr, err := json.Marshal(StdResponse{Code: code, Error: error, Message: message})

	if err != nil {
		log.Println(err)
		return
	}

	var res = http.Response{Proto: "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Close:      true,
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(jsonifiedStr))}

	err = res.Write(client)
	if err != nil {
		return
	}
}
