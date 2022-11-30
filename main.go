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

var filePath = os.Getenv("FILE_PATH")

func main() {
	port := os.Args[1]
	startServer(port)
}

/*
*
Start the webserver, and begin listening for incoming clients.
*/
func startServer(port string) {

	var bindAddr string

	/*
		If environment variables above are not set, it means we are running locally, set the file path and bind address accordingly
	*/

	if filePath == "" {
		filePath = "/Users/karthik/Public/tmp/"
		bindAddr = "localhost"
	} else {
		bindAddr = "0.0.0.0"
	}

	log.Println("Server starting on port: " + port)
	server, err := net.Listen("tcp", bindAddr+":"+port)
	if err != nil {
		log.Fatal("Error starting server: " + err.Error())
	}
	weightedSem := sem.NewWeighted(10) //Semaphore to ensure no more than 10 concurrent clients.
	for {
		semErr := weightedSem.Acquire(context.Background(), 1) //Grab one per client
		client, err := server.Accept()
		if err != nil || semErr != nil {
			log.Println("Error accepting new client: " + err.Error())
			return
		} else {
			go processClient(client, weightedSem)
		}
	}
}

/*
*
Function to handle incoming requests.
This function is always called on separate go-routines, and handles all the client data.
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
		log.Println("Error building request: ", err)
		return
	}

	log.Println("URL path: " + req.URL.Path)
	log.Println("Request method: " + req.Method)

	url := req.URL.Path
	urlSlices := strings.Split(url, "/")
	resourceName := urlSlices[len(urlSlices)-1]

	/*
		This if statement distinguishes between the different HTTP methods, and handles each appropriate
	*/

	/*------------------------GET---------------------*/
	if req.Method == "GET" { // handle get

		getHandler(resourceName, client, req)

		sendResponse(http.StatusNotFound, true, "Please enter the right file name or upload a file", client)

		/*------------------------POST---------------------*/
	} else if req.Method == "POST" { //handle post

		postHandler(req, client)

	} else {
		sendResponse(http.StatusNotImplemented, true, "Only GET and POST methods are supported", client)
	}
}

func getHandler(resourceName string, client net.Conn, req *http.Request) {
	if resourceName != "" {

		sliceFileName := strings.Split(resourceName, ".")
		ext := sliceFileName[len(sliceFileName)-1]

		allowedExts := [6]string{"html", "txt", "gif", "jpeg", "jpg", "css"}

		i := -1
		for i < len(allowedExts) {
			i++
			if i == len(allowedExts) {
				sendResponse(http.StatusNotImplemented, true, ext+" is not allowed", client)
				return
			}
			if ext == allowedExts[i] {
				break
			}
		}

		file, err := os.ReadFile(filePath + "" + resourceName + "")
		mimeType := http.DetectContentType(file) // Get Content-Type for adding it to the header

		if err != nil {
			log.Println(err)
			sendResponse(400, true, err.Error(), client)
			return
		} else {
			req.Header = make(http.Header)
			req.Header.Set("Content-Type", mimeType) // Add the above Content-Type to the header
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
}

func postHandler(req *http.Request, client net.Conn) {

	retVal := multipartUpload(req)

	var respCode = 0
	var msg = "File uploaded"

	if retVal {
		respCode = http.StatusCreated
	} else {
		respCode = http.StatusBadRequest
		msg = "Wrong File format or incomplete upload"
	}

	sendResponse(respCode, !retVal, msg, client)
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

		sliceFileName := strings.Split(part.FileName(), ".")
		ext := sliceFileName[len(sliceFileName)-1]

		allowedExts := [6]string{"html", "txt", "gif", "jpeg", "jpg", "css"}

		i := -1
		for i < len(allowedExts) {
			i++
			if i == len(allowedExts) { // If not in allowed extensions
				return false // This will throw a 400 Bad request error
			}
			if ext == allowedExts[i] {
				break
			}
		}

		d, err := os.Create(filePath + part.FileName()) // Create a file in the path with given file name on the filesystem
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
		_, err = io.Copy(d, part) // Copy the file from request to the newly created file on the the file system
		if err != nil {
			return false
		}
	}
	return true // This will throw a 201 Created Response
}

/*
*
Function for sending responses to the client.
*/
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
