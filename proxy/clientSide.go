package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"lab1DS/sem"
	"log"
	"net"
	"net/http"
	"strings"
)

func startServer(port string) {
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
	readLen, errOne := client.Read(tmp)
	print("First read: ", readLen)
	buffer.Write(tmp[:readLen])
	defer client.Close()
	for readLen == bufSize {
		tot += readLen
		print("read: ", readLen)
		if errOne != nil {
			if errOne == io.EOF {
				print("EOF REACHED")
				break
			}
			log.Println(errOne)
			fmt.Println("Error reading request: " + errOne.Error())
			client.Close()
			return
		}

		fmt.Println("Error reading request: " + errOne.Error())
		buffer.Write(tmp[:readLen])
		readLen, errOne = client.Read(tmp)
	}
	if tot > bufSize {
		buffer.Write([]byte("\n\r")) //Fucking golang
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
	print(urlSlices)

	if req.Method == "GET" { // handle get
		// call karthik
	} else {
		res := http.Response{StatusCode: 501, Close: true}
		res.Write(client)
	}

}
