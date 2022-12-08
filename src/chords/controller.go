package main

import (
	"encoding/json"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"sort"
	"strconv"
	"time"
)

var predecessors PeerList
var successors PeerList
var fingerTable PeerList
var OWN_ID string
var neighborsLen int = 3

//func testSetup() {
//	OWN_ID = "12778624"
//}

/*
1. -a <String> = The IP address that the Chord client will bind to, as well as advertise to other nodes. Represented as
					an ASCII string (e.g., 128.8.126.63). Must be specified.
2. -p <Number> = The port that the Chord client will bind to and listen on. Represented as a base-10 integer. Must be specified.
3. --ja <String> = The IP address of the machine running a Chord node. The Chord client will join this node’s ring.
					Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified.
4. --jp <Number> = The port that an existing Chord node is bound to and listening on. The Chord client will join this node’s ring.
					Represented as a base-10 integer. Must be specified if --ja is specified.
5. --ts <Number> = The time in milliseconds between invocations of ‘stabilize’.
					Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].
6. --tff <Number> = The time in milliseconds between invocations of ‘fix fingers’. Represented as a base-10 integer.
					Must be specified, with a value in the range of [1,60000].
7. --tcp <Number> = The time in milliseconds between invocations of ‘check predecessor’.
					Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].
8. -r <Number> = The number of successors maintained by the Chord client. Represented as a base-10 integer.
					Must be specified, with a value in the range of [1,32].
9. -i <String> = The identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s
					IP address and port number. Represented as a string of 40 characters matching [0-9a-fA-F]. Optional parameter.
*/

func main() {
	//Parse args

	ip := "127.0.0.1" // defaults to localhost.
	port := "12323"   //default port
	maintenanceTime := 30000 * time.Millisecond
	joinIp := ""
	joinPort := ""

	for i := 1; i < len(os.Args)-1; i++ {

		switch os.Args[i] {
		case "-a":
			ip = os.Args[i+1]
			println("Binding to IP: ", os.Args[i+1])
			break
		case "-p":
			port = os.Args[i+1]
			println("Binding Port: ", os.Args[i+1])
			break
		case "--ja": //Join ip
			println("Joining IP: ", os.Args[i+1])
			joinIp = os.Args[i+1]
			break
		case "--jp": //join port
			println("Joining Port: ", os.Args[i+1])
			joinPort = os.Args[i+1]
			break
		case "--ts":
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			println("Stabilizing Time: ", os.Args[i+1])
			break
		case "--tff":
			println(os.Args[i+1])
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			break
		case "--tcp":
			println(os.Args[i+1])
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			break
		case "-i":
			println("Overridden ID: ", os.Args[i+1])
			OWN_ID = os.Args[i+1]
			break
		default:
			continue
		}
	}

	//testSetup()

	if OWN_ID == "" {
		log.Println("Setting random ID.")
		OWN_ID = "172b17"

	}

	print("Start\n")

	createNetworkInstance(ip, port)
	if joinIp != "" && joinPort != "" {
		log.Println("Attempting to join: " + ip + ":" + port)
		join(joinIp, joinPort)

	}

	maintenanceLoop(maintenanceTime)
}

func getOwnerStruct() Peer {
	return Peer{Ip: OWN_IP, Port: OWN_PORT, ID: OWN_ID}
}

func checkNeighbors() {
	for i := 0; i < predecessors.Len(); i++ {
		notify(predecessors[i])
	}
	for i := 0; i < len(successors); i++ {
		notify(successors[i])
	}
}

func maintainFingers() {

	log.Println("Checking finger-table. Own ID: ", OWN_ID)
	var slider int64 = 1
	index := 0
	for ; index < 64; index++ {

		curSearch, _ := new(big.Int).SetString(OWN_ID, 16)
		curSearch.Add(curSearch, new(big.Int).SetInt64(slider<<index))
		searchTerm := curSearch.Text(16)

		startPoint := fingerSearch(searchTerm)
		if startPoint != nil {
			var tmpVars []string
			tmpVars = append(tmpVars, searchTerm)
			bytes, err := json.Marshal(MessageType{Owner: getOwnerStruct(), Action: "search", Vars: tmpVars})
			if err != nil {
				log.Println("ERROR MARSHALLING SEARCH IN FINGER MAINTENANCE: ", err.Error())
			}
			result := search(bytes, *startPoint)
			if index > len(fingerTable) || (result.ID > fingerTable[index].ID && result.ID < searchTerm) {
				fingerTable[index] = result
			}
		}
	}
}

func maintenanceLoop(mTime time.Duration) {
	for {

		checkNeighbors()
		maintainFingers()
		time.Sleep(mTime)
	}
}

func fingerSearch(id string) *Peer {
	var foundPeer *Peer = nil
	searchTerm, _ := new(big.Int).SetString(id, 16)
	i := 0
	for ; i < fingerTable.Len(); i++ {
		tmp, noProbs := new(big.Int).SetString(fingerTable[i].ID, 16)
		if noProbs && tmp.Cmp(searchTerm) > 0 {
			foundPeer = &fingerTable[i]
		} else if !noProbs {
			log.Println("PROBS!")
		}
	}
	return foundPeer
}

func searchResponse(message MessageType, conn net.Conn) {
	log.Println("Building search response.")
	log.Println("Vars: ", message.Vars, "LEN: ", len(message.Vars))
	//message.Vars = []string(message.Vars)
	var foundPeer Peer

	if message.Vars[0] <= OWN_ID {
		log.Println("Peer found itself as successor. Sending self as search response")
		foundPeer = getOwnerStruct()
	} else if len(successors) > 1 && message.Vars[0] < successors[0].ID {
		foundPeer = successors[0]
	} else if len(successors) > 2 && message.Vars[0] < successors[1].ID {
		foundPeer = successors[1]
	} else if len(successors) > 1 && message.Vars[0] < successors[2].ID {
		foundPeer = successors[2]
	} else {
		tmpPeer := fingerSearch(message.Vars[0])
		if tmpPeer != nil {
			foundPeer = *tmpPeer
		} else {
			foundPeer = getOwnerStruct()
		}
	}

	peerBytes, err := json.Marshal(foundPeer)
	if err != nil {
		log.Println("ERROR MARSHALLING SEARCH RESPONSE: ", err.Error())
	}
	var vars []string
	vars = append(vars, string(peerBytes))
	response, err := json.Marshal(MessageType{
		Action: "searchResponse",
		Owner:  getOwnerStruct(),
		Vars:   vars,
	})

	if err != nil {
		log.Println("ERROR MARSHALLING SEARCH RESPONSE MESSAGE: ", err.Error())
		return
	}

	sendToPeer(conn, response)
}

func respondToNotify(message MessageType, conn net.Conn) {

	newList := new(PeerList)

	newList.Append(message.Owner)
	refreshNeighbours(newList)

	var resList []string

	for i := 0; i < predecessors.Len(); i++ {
		preBytes, _ := json.Marshal(predecessors[i])
		resList = append(resList, string(preBytes))
	}

	response := MessageType{Action: "notifyResponse", Owner: getOwnerStruct(), Vars: resList}

	jsonres, _ := json.Marshal(response)
	sendToPeer(conn, jsonres)

}

func handleIncoming(message MessageType, conn net.Conn) {
	switch message.Action {
	case "notify":
		respondToNotify(message, conn)
		break
	case "search":
		searchResponse(message, conn)
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

	println("FINISHED JOIN")

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
	//sucConn.Close()
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
	sucConn.Close()
	refreshNeighbours(&neighbors)
}

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
		println("NEWLIST len : ", len(newList))
		tempPredecessors := newList[0:i]
		tempSuccessors := newList[i:len(newList)]

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

		successors = tempSuccessors[:int(math.Min(float64(tempSuccessors.Len()), float64(neighborsLen)))]
		predecessors = tempPredecessors[:int(math.Min(float64(tempPredecessors.Len()), float64(neighborsLen)))] //TODO: CHECK INDEXING.

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

func (a PeerList) Append(b Peer) {
	a = append(a, b)
}
