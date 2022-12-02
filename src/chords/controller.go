package main

import (
	"encoding/json"
	"log"
	"time"
)

var predecessors [3]Peer
var successors [3]Peer
var fingerTable [161]Peer
var OWN_ID string

func main() {
	//Parse args
	maintanenceTime := 5000 * time.Millisecond
	print("Start")

	createNetworkInstance("127.0.0.1", "12323")
	maintanenceLoop(maintanenceTime)
}

func maintanenceLoop(mTime time.Duration) {

	for {
		time.Sleep(mTime)
	}
}

func join(ip string, port string) {

	message, err := json.Marshal(
		MessageType{
			Action: "search",
			Owner: Peer{
				Ip:   OWN_IP,
				Port: OWN_PORT,
				ID:   OWN_ID,
			},
			Vars: []string{OWN_ID},
		})
	if err != nil {
		log.Println("ERROR MARSHALLING JOIN MESSAGE: ", err.Error())
	}
	search(message, Peer{Ip: ip, Port: port})

}
