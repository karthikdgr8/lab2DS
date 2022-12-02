package main

import (
	"encoding/json"
	"log"
	"sort"
	"time"
)

var predecessors []Peer
var successors []Peer
var fingerTable [161]Peer
var OWN_ID string

func main() {
	//Parse args
	maintanenceTime := 5000 * time.Millisecond
	print("Start")

	createNetworkInstance("127.0.0.1", "12323")
	maintanenceLoop(maintanenceTime)
}

func getOwnerStruct() Peer {
	return Peer{Ip: OWN_IP, Port: OWN_PORT, ID: OWN_ID}
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
	succ := search(message, Peer{Ip: ip, Port: port})
	notify(succ)

}

/*
	Function to notify any peer of our own existence. This generalized starts by sending a "notify"
	message to the relevant peer. This other peer will then, based on its own logic update its relevant tables
	and respond with its closest neighbors in the "var" field.
	The structure of the response will be the following.

	var: [
		{first predecessor},
		{second predecessor},
		{third predecessor},
		{first successor},
		{second successor},
		{third successor}
	]

	The order here is however arbitrary, as the list will be sorted before use.
*/

func notify(peer Peer) {
	sucConn := connectToPeer(peer.Ip, peer.Port)
	message, err := json.Marshal(
		MessageType{
			Owner:  getOwnerStruct(),
			Action: "notify"})

	if err != nil {
		log.Println("ERROR MARSHALLING NOTIFY MESSAGE: ", err.Error())
	}
	sendToPeer(sucConn, message)
	response := MessageType{}
	err = json.Unmarshal(listenForData(sucConn), &response)
	rawNeighbors := response.Vars
	var neighbors PeerList
	for i := 0; i < len(rawNeighbors); i++ {
		tmpPeer := Peer{}
		err = json.Unmarshal([]byte(rawNeighbors[i]), &tmpPeer)
		if err != nil {
			log.Println("ERROR UNMARSHALLING NEIGHBOR LIST: ", err.Error())
			break
		}
		neighbors = append(neighbors, tmpPeer)
	}

	neighbors = append(neighbors, peer) // Append the notified peer itself.
	sort.Sort(neighbors)
}

func refreshNeighbours(newList *PeerList) {
	if newList != nil {
		newList := *newList // Dereference into overshadowing variable.
		if len(successors) > 0 {

		}
		//Find our index in the list.
		i := 0
		for i = 0; i < newList.Len(); i++ {
			if newList[i].ID > OWN_ID {
				break
			}
		}
		tmpPredecessors := newList[:i-1]
		tempSuccessors := newList[2:]

		for i = 0; i < 3; i++ {
			if i < tempSuccessors.Len() {
				//TODO: implement.
			}
			if i < tmpPredecessors.Len() {

			}
		}
	}
}

/*Typedef needed to aid in sorting.*/
type PeerList []Peer

/*Sort.Interface methods to aid with sorting.*/
func (a PeerList) Len() int {
	return len(a)
}
func (a PeerList) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}
func (a PeerList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
