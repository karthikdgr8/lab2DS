package control

import (
	"bytes"
	"io"
	"lab1DS/src/peerNet"
	"log"
	"net"
	"os"
)

var filePath = ""

func storeFile(fileName string, conn net.Conn) {

	data := peerNet.ListenForData(conn)

	dataReader := bytes.NewReader(data)

	d, err := os.Create(filePath + fileName) // Create a file in the path with given file name on the filesystem
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

	_, err = io.Copy(d, dataReader) // Copy the file from request to the newly created file on the the file system
	if err != nil {
		return
	}
}

func searchFile(fileName string) []byte {

	file, err := os.ReadFile(filePath + "" + fileName + "")
	if err != nil {
		log.Println("Error reading file or file not found" + err.Error())
		return nil
	}

	return file
}
