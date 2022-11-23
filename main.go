package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"lab1DS/sem"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
)

type StdResponse struct {
	Code    int
	Error   bool
	Message string
}

const filePath = "/Users/karthik/Public/tmp/"

func main() {
	port := os.Args[1]
	startServer(port)

}

func startServer(port string) {
	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	weightedSem := sem.NewWeighted(10)
	for {
		semErr := weightedSem.Acquire(context.Background(), 1)
		client, err := server.Accept()
		if err != nil || semErr != nil {
			log.Println("Error accepting new client: " + err.Error())
			return
		} else {
			go processClient(client, weightedSem)
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
	resourceName := urlSlices[len(urlSlices)-1]
	println("resourceName", resourceName)

	if req.Method == "GET" { // handle get
		if resourceName != "" {
			println("In Here")
			file, err := os.ReadFile(filePath + "" + resourceName + "")
			mimeType := http.DetectContentType(file)

			if err != nil {
				log.Println(err)
				sendResponse(400, true, err.Error(), client)
				return
			} else {
				req.Header = make(http.Header)
				req.Header.Set("Content-Type", mimeType)
				var res = http.Response{Close: true,
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(file)),
					Header:     req.Header}

				err := res.Write(client)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}

		sendResponse(http.StatusNotFound, true, "Please enter the right file name or upload a file", client)

	} else if req.Method == "POST" { //handle post

		retVal := multipartUpload(req)

		var respCode = 0

		if retVal {
			respCode = http.StatusCreated
		} else {
			respCode = http.StatusBadRequest
		}

		var res = http.Response{Close: true, StatusCode: respCode}

		err := res.Write(client)
		if err != nil {
			log.Println(err)
			return
		}
		//sendResponse(respCode, false, "File uploaded", client)

	} else {
		sendResponse(http.StatusNotImplemented, true, "Only GET and POST methods are supported", client)
	}
}

func multipartUpload(req *http.Request) bool {

	reader, err := req.MultipartReader()
	if err != nil {
		log.Println(err)
		return false
	}
	for {

		part, err := reader.NextPart()

		if err == io.EOF {
			break
		} else if err != nil {
			break
		}

		defer func(part *multipart.Part) {
			err := part.Close()
			if err != nil {
				log.Println(err)
				return
			}
		}(part)
		if part.FileName() == "" {
			continue
		}

		d, err := os.Create(filePath + part.FileName())
		if err != nil {
			log.Println(err)
			return false
		}
		defer func(d *os.File) {
			err := d.Close()
			if err != nil {
				return
			}
		}(d)
		_, err = io.Copy(d, part)
		if err != nil {
			return true
		}
	}
	return true
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
		return
	}
}
