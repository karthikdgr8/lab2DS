package main

import (
	"encoding/json"
	"github.com/holiman/uint256"
	"log"
	"math/big"
	"sort"
	"time"
)

var predecessors PeerList
var successors PeerList
var fingerTable PeerList
var OWN_ID string

func main() {
	//Parse args
	maintanenceTime := 5000 * time.Millisecond
	print("Start")

	createNetworkInstance("127.0.0.1", "12323")
	maintenanceLoop(maintanenceTime)
}

func getOwnerStruct() Peer {
	return Peer{Ip: OWN_IP, Port: OWN_PORT, ID: OWN_ID}
}

func maintenanceLoop(mTime time.Duration) {

	for {
		time.Sleep(mTime)
	}
}

func fingerSearch (id string) Peer{
	var foundPeer *Peer
 	searchID := new(big.Int)
	searchID.SetString(id, 16)
	searchID.Abs(searchID)
	ownId := new(big.Int)
	ownId.SetString(OWN_ID, 16)
	ownId.Abs(ownId)
	fingerInt := *ownId
	index := 0
	var shifter int64
	mytest := uint256.NewInt(2)
	adder := new(big.Int)
	for searchID.Cmp(&fingerInt) < 0 && index < 162 {
		fingerInt = *ownId
		if index < 64{
			shifter = 1
			fingerInt.Add(&fingerInt, new(big.Int).SetInt64(shifter<<index)) //int64( math.Pow(2, index))))
		} else{
			if(index < 64 && index < )

		}

		shifter = 1 << (index % 64)
		if index > 64 && index < 128 {
			adder.MulRange()
			adder.Mul(shifter, 1<<63)
		} else if index > 128 {

		} else {

		}

		index++
	}
	foundPeer = &fingerTable[index-1]
	return *foundPeer
}

func searchResponse(message MessageType) {
	var foundPeer Peer
	if message.Vars[0] < OWN_ID {
		foundPeer = getOwnerStruct()
	} else if message.Vars[0] < successors[0].ID {
		foundPeer = successors[0]
	} else if message.Vars[0] < successors[1].ID {
		foundPeer = successors[1]
	} else if message.Vars[0] < successors[2].ID {
		foundPeer = successors[2]
	} else {

	}

	peerBytes, err := json.Marshal(foundPeer)
	if err != nil {
		log.Println("ERROR MARSHALLING SEARCH RESPONSE: ", err.Error())
	}
	var vars []string
	vars[0] = string(peerBytes)

	response, err := json.Marshal(MessageType{
		Action: "searchResponse",
		Owner:  getOwnerStruct(),
		Vars:   vars,
	})

	if err != nil {
		log.Println("ERROR MARSHALLING SEARCH RESPONSE MESSAGE: ", err.Error())
		return
	}
	sendToPeer(connectToPeer(message.Owner.Ip, message.Owner.Port), response)
}

func handleIncoming(message MessageType) {
	switch message.Action {
	case "notify":
		break
	case "search":
		searchResponse(message)
		break
	case "put":
		break
	case "fetch":
		break

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
	Function to notify any peer of our own existence. This generalized function starts by sending a "notify"
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
	refreshNeighbours(&neighbors)
}

// input must be sorted.
func refreshNeighbours(newList *PeerList) {
	if newList != nil {
		newList := *newList // Dereference into overshadowing variable.
		//Find our index in the list.
		i := 0
		for i = 0; i < newList.Len(); i++ {
			if newList[i].ID > OWN_ID {
				break
			}
		}
		tempPredecessors := newList[0 : i-1]
		tempSuccessors := newList[i : len(newList)-1]

		for i := 0; i < 3; i++ {
			if i < predecessors.Len() {
				tempPredecessors = append(tempPredecessors, predecessors[i])
			}
			if i < successors.Len() {
				tempSuccessors = append(tempSuccessors, successors[i])
			}
		}

		doNotify := !sort.IsSorted(tempPredecessors) || !sort.IsSorted(successors)
		sort.Sort(tempSuccessors)
		sort.Sort(tempPredecessors)

		successors = tempSuccessors[:3]
		predecessors = tempPredecessors[:3] //TODO: CHECK INDEXING.

		if doNotify {
			for i := 0; i < successors.Len(); i++ {
				notify(successors[i])
			}
			for i := 0; i < predecessors.Len(); i++ {
				notify(predecessors[i])
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
