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
	"strings"
)

type StdResponse struct {
	Code    int
	Error   bool
	Message string
}

func startProxy(port string) {
	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	weigtedSem := sem.NewWeighted(10)
	for {
		semErr := weigtedSem.Acquire(context.Background(), 1)
		client, err := server.Accept()
		if err != nil || semErr != nil {
			log.Println("Error accepting new client: " + err.Error())
			sendResponse(http.StatusInternalServerError, true, err.Error(), client)
			return
		} else {
			go processClient(client, weigtedSem)
		}
	}
}

func processClient(client net.Conn, weighted *sem.Weighted) {
	log.Println("New client accepted,", client.RemoteAddr())
	defer weighted.Release(1)
	bufSize := 256
	buf := make([]byte, 0) // do not use.
	tmp := make([]byte, bufSize)
	buffer := bytes.NewBuffer(buf)
	tot := 0
	defer client.Close()
	for {
		readLen, err := client.Read(tmp)
		if err != nil {
			if err == io.EOF {
				print("EOF REACHED")
				break
			}
			log.Println("Error inside", err)
			client.Close()
		}

		buffer.Write(tmp[:readLen])
		tot += readLen
		if readLen < 256 && tot != 0 {
			break
		}
	}

	reader := bufio.NewReader(buffer)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Error building request: ", err)
		sendResponse(http.StatusInternalServerError, true, err.Error(), client)
		return
	}

	log.Println("URL path: " + req.URL.Path)
	log.Println("Request method: " + req.Method)

	url := req.URL.Path
	urlSlices := strings.Split(url, "/")
	print("urlSlices", urlSlices)

	if req.Method == "GET" { // handle get

		resp := callServer(req)

		err := resp.Write(client)
		if err != nil {
			log.Println(err)
		}
	} else {
		sendResponse(http.StatusNotImplemented, true, err.Error(), client)
	}

}

func sendResponse(code int, error bool, message string, client net.Conn) {
	jsonifiedStr, err := json.Marshal(StdResponse{Code: code, Error: error, Message: message})

	if err != nil {
		log.Println(err)
		return
	}

	var res = http.Response{Close: true,
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(jsonifiedStr))}

	err = res.Write(client)
	if err != nil {
		log.Println(err)
		return
	}
}
