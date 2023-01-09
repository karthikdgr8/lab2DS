package peerNet

import (
	"bufio"
	"encoding/base64"
	"lab1DS/src/becauseGO"
	"log"
	"net"
)

var OWN_IP string
var OWN_PORT string
var DELIM = 0x1F
var controller becauseGO.Callback

func CreateNetworkInstance(ip string, port string, callback becauseGO.Callback) {
	controller = callback
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
		buf, err := base64.URLEncoding.DecodeString(string(data))
		data = buf
		if err != nil {
			log.Panic("ERROR DECODING B64 message")
		}
	}

	if err != nil {
		log.Println("ERROR READING DATA: " + err.Error())
		return nil
	}
	//println("Read: ", string(data))
	return data
}

func ConnectToPeer(ip string, port string) net.Conn {
	log.Println("Dialing out.")
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Println("ERROR ESTABLISHING CLIENT TCP: " + err.Error())
		return nil
	}
	return conn
}

func SendToPeer(conn net.Conn, data []byte) {

	println("Sending : ", string(data))

	var sendBuf = base64.URLEncoding.EncodeToString(data)
	_, err := conn.Write(append([]byte(sendBuf), byte(DELIM)))
	if err != nil {
		log.Println("ERROR SENDING DATA: " + err.Error())
	}
}

// Handle new client.
func handleNewConnection(conn net.Conn) {

	//data := ListenForData(conn)
	//HandleIncoming(data, conn)
	controller.Callback(conn)
	//conn.Close()
}
