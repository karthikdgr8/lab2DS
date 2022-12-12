package peerNet

import (
	"bufio"
	"encoding/json"
	"lab1DS/src/protocol"
	"lab1DS/src/ring"
	"log"
	"net"
)

func PublicFunc() {

}

var OWN_IP string
var OWN_PORT string
var DELIM = 0x1F

func CreateNetworkInstance(ip string, port string) {
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

func ListenForData(conn net.Conn) []byte {
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

func ConnectToPeer(ip string, port string) net.Conn {
	log.Println("Dialing out.")
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Println("ERROR ESTABLISHING CLINET TCP: " + err.Error())
		return nil
	}
	return conn
}

func SendToPeer(conn net.Conn, data []byte) {
	println("Sending : ", string(data))
	_, err := conn.Write(append(data, byte(DELIM)))
	if err != nil {
		log.Println("ERROR SENDING DATA: " + err.Error())
	}
}

func unmarshalPeer(data string) *ring.Peer {
	ret := ring.Peer{}
	err := json.Unmarshal([]byte(data), &ret)
	if err != nil {
		log.Println("ERROR UNMARSHALLING PEER. Input: "+data, "Error: "+err.Error())
		return nil
	}
	return &ret
}

// Handle new client.
func handleNewConnection(conn net.Conn) {
	data := ListenForData(conn)
	message := protocol.MessageType{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err.Error())
	}
	//handleIncoming(message, conn)
	conn.Close()

}

func Search(message []byte, startPoint ring.Peer) ring.Peer {
	log.Println("Searching..")
	destination := ring.Peer{}
	for {
		conn := ConnectToPeer(startPoint.Ip, startPoint.Port)
		log.Println("Connected to search entrypoint. Sending search: ", string(message))
		SendToPeer(conn, message)
		//_, err := conn.Write(message)
		response := ListenForData(conn)
		parsedResponse := protocol.MessageType{}
		conn.Close()

		err := json.Unmarshal(response, &parsedResponse)
		if err != nil {
			log.Println("ERROR UNMARSHALLING SEARCH RESPONSE: ", err.Error())
			return ring.Peer{}
		}
		err = json.Unmarshal([]byte(parsedResponse.Vars[0]), &destination)
		if err != nil {
			log.Println("ERROR UNMARSHALING DESTINATION PEER: ", err.Error())
			return ring.Peer{}
		}
		if destination.ID == startPoint.ID {
			return destination
		} else {
			startPoint = destination // Next entry for search
		}
	}
}
