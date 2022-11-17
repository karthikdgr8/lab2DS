package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
)

type stdResponse struct {
	code    int
	error   bool
	message string
}

func main() {
	port := os.Args[1]
	startServer(port)
}

func startServer(port string) {
	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}

	for {
		client, err := server.Accept()
		if err != nil {
			log.Println("Error accepting new client: " + err.Error())
			return
		} else {
			go processClient(client)
		}
	}
}

func processClient(client net.Conn) {
	log.Println("New client accepted,", client.RemoteAddr())
	buf := make([]byte, 0) // do not use.
	tmp := make([]byte, 10048)
	buffer := bytes.NewBuffer(buf)
	readLen, err := client.Read(tmp)
	if err != nil {
<<<<<<< Updated upstream
		log.Println(err)
=======
		fmt.Println("Error reading request: " + err.Error())
		client.Close()
>>>>>>> Stashed changes
		return
	}
	buffer.Write(tmp[:readLen])
	reader := bufio.NewReader(buffer)
	req, err := http.ReadRequest(reader)

<<<<<<< Updated upstream
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("URL path: " + req.URL.Path)
	log.Println("Request method: " + req.Method)

=======
	url := req.URL.Path
	urlSlices := strings.Split(url, "/")
	resourceName := urlSlices[len(urlSlices)-1]
	fmt.Println("Got request: " + req.Method)
>>>>>>> Stashed changes
	if req.Method == "GET" { // handle get
		var res = http.Response{Close: true, StatusCode: 200
		
		}

	} else if req.Method == "POST" { //handle post

		multipartUpload(req)

		var res = http.Response{Close: true, StatusCode: 200}

		err := res.Write(client)
		if err != nil {
			log.Println(err)
			return
		}
		req.Close = true

	} else {
		var res = http.Response{Close: true, StatusCode: 501}

<<<<<<< Updated upstream
		err := res.Write(client)
		if err != nil {
			log.Println(err)
			return
		}
=======
		res.Write(client)
>>>>>>> Stashed changes

		req.Close = true
	}
}

func multipartUpload(req *http.Request) {

	reader, err := req.MultipartReader()
	if err != nil {
		log.Println(err)
		return
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

		d, err := os.Create("/Users/karthik/Public/tmp/" + part.FileName())
		if err != nil {
			log.Println(err)
			return
		}
		defer func(d *os.File) {
			err := d.Close()
			if err != nil {
				return
			}
		}(d)
		_, err = io.Copy(d, part)
		if err != nil {
			return
		}
	}
}

func sendResponse(code int, error bool, message string) []byte {
	jsonifiedStr, err := json.Marshal(stdResponse{code: code, error: error, message: message})
	if err != nil {
		log.Println(err)
		return nil
	}

	return jsonifiedStr
}
