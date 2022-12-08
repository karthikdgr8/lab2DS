package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

type Peer struct {
	Ip   string
	Port string
	ID   string
	//Connection net.Conn `json:"-"`
	//SendSem    sem.Weighted `json:"-"`
}

type MessageType struct {
	Action string
	Owner  Peer
	Vars   []string
}

var OWN_IP string
var OWN_PORT string
var DELIM = 0x1F

func createNetworkInstance(ip string, port string) {
	OWN_IP = ip
	OWN_PORT = port
	startNetworkInstance()
}

func startNetworkInstance() {
	log.Println("STARTING TCP-SERVER")
	server, err := net.Listen("tcp", OWN_IP+":"+OWN_PORT)
	if err != nil {
		log.Println("Error: " + err.Error())
	}
	go serverRoutine(server)

}

func serverRoutine(listener net.Listener) {
	for {
		client, err := listener.Accept()
		log.Println("New client accepted")
		if err != nil {
			log.Println("ERROR ACCEPTING CLIENT: " + err.Error())
		} else {
			go handleNewConnection(client)
		}
	}
}

func listenForData(conn net.Conn) []byte {
	log.Println("Reading from client.")
	connBuf := bufio.NewReader(conn)
	data, err := connBuf.ReadBytes(byte(DELIM))
	if len(data) > 1 {
		data = data[:len(data)-1]
	}

	if err != nil {
		log.Println("ERROR READING DATA: " + err.Error())
		return nil
	}
	println("Read: ", string(data))
	return data
}

func connectToPeer(ip string, port string) net.Conn {
	log.Println("Dialing out.")
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Println("ERROR ESTABLISHING CLINET TCP: " + err.Error())
		return nil
	}
	return conn
}

func sendToPeer(conn net.Conn, data []byte) {
	println("Sending : ", string(data))
	_, err := conn.Write(append(data, byte(DELIM)))
	if err != nil {
		log.Println("ERROR SENDING DATA: " + err.Error())
	}
}

func unmarshalMessage(data []byte) *MessageType {
	message := MessageType{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err.Error())
		return nil
	}
	return &message
}

func unmarshalPeer(data string) *Peer {
	ret := Peer{}
	err := json.Unmarshal([]byte(data), &ret)
	if err != nil {
		log.Println("ERROR UNMARSHALLING PEER. Input: "+data, "Error: "+err.Error())
		return nil
	}
	return &ret
}

// Handle new client.
func handleNewConnection(conn net.Conn) {
	data := listenForData(conn)
	message := MessageType{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err.Error())
	}
	handleIncoming(message, conn)
	conn.Close()

}

func search(message []byte, startPoint Peer) Peer {
	log.Println("Searching..")
	destination := Peer{}
	for {
		conn := connectToPeer(startPoint.Ip, startPoint.Port)
		log.Println("Connected to search entrypoint. Sending search: ", string(message))
		sendToPeer(conn, message)
		//_, err := conn.Write(message)
		response := listenForData(conn)
		parsedResponse := MessageType{}
		conn.Close()

		err := json.Unmarshal(response, &parsedResponse)
		if err != nil {
			log.Println("ERROR UNMARSHALLING SEARCH RESPONSE: ", err.Error())
			return Peer{}
		}
		err = json.Unmarshal([]byte(parsedResponse.Vars[0]), &destination)
		if err != nil {
			log.Println("ERROR UNMARSHALING DESTINATION PEER: ", err.Error())
			return Peer{}
		}
		if destination.ID == startPoint.ID {
			return destination
		} else {
			startPoint = destination // Next entry for search
		}
	}
}
