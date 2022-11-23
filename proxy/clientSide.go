package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"lab1DS/sem"
	"log"
	"net"
	"net/http"
	"strings"
)

func startProxy(port string) {
	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	weigtedSem := sem.NewWeighted(10)
	for {
		semErr := weigtedSem.Acquire(context.Background(), 1)
		client, err := server.Accept()
		if err != nil || semErr != nil {
			log.Println("Error accepting new client: " + err.Error())
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
		return
	}

	log.Println("URL path: " + req.URL.Path)
	log.Println("Request method: " + req.Method)

	url := req.URL.Path
	urlSlices := strings.Split(url, "/")
	print("urlSlices", urlSlices)

	if req.Method == "GET" { // handle get
		print("Calling server\n")
		resp := callServer(req)
		print("Server responded: \n")
		err := resp.Write(client)
		if err != nil {
			log.Println(err)
		}
	} else {
		res := http.Response{StatusCode: 501, Close: true}
		err := res.Write(client)
		if err != nil {
			log.Println(err)
			return
		}
	}

}
