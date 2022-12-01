package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

type Client struct {
	Ip   string
	Port string
	ID   string
}

type MessageType struct {
	Action string
	Owner  Client
	Vars   []string
}

var IP string
var PORT string
var DELIM = 0x1F

func createNetworkInstance(ip string, port string) {
	IP = ip
	PORT = port
	startNetworkInstance()
}

func startNetworkInstance() {
	log.Println("STARTING TCP-SERVER")
	server, err := net.Listen("tcp", IP+":"+PORT)
	if err != nil {
		log.Println("Error: " + err.Error())
	}
	go serverRoutine(server)

}

func serverRoutine(listener net.Listener) {
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println("ERROR ACCEPTING CLIENT: " + err.Error())
		} else {
			go handleNewConnection(client)
		}
	}
}

func listenForData(conn net.Conn) []byte {
	log.Println("New client accepted")
	connBuf := bufio.NewReader(conn)
	data, err := connBuf.ReadBytes(byte(DELIM))
	data = data[:len(data)-1]
	if err != nil {
		log.Println("ERROR READING DATA: " + err.Error())
		return nil
	}
	return data
}

func connectToPeer(ip string, port string) net.Conn {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Println("ERROR ESTABLISHING CLINET TCP: " + err.Error())
		return nil
	}
	return conn
}

func clientSend(conn net.Conn, data []byte) {
	_, err := conn.Write(append(data, byte(DELIM)))
	if err != nil {
		log.Println("ERROR SENDING DATA: " + err.Error())
	}
}

func unmarshalMessage(data []byte) MessageType {
	message := MessageType{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err.Error())
		return MessageType{}
	}
	return message
}

// Handle new client.
func handleNewConnection(conn net.Conn) {
	data := listenForData(conn)
	message := MessageType{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err.Error())
	}

}
