package control

import (
	"encoding/json"
	"lab1DS/src/becauseGO"
	"lab1DS/src/peerNet"
	"lab1DS/src/ring"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var OWN_ID string
var RING *ring.Ring
var NEIGH_LEN int

func StartUp(ip, port string, neigborsLen int, maintenanceTime time.Duration, ownID, joinIp, joinPort string) {
	OWN_ID = ownID
	netCallback := becauseGO.Callback{Callback: HandleIncoming}
	getKeyFromFile()
	peerNet.CreateNetworkInstance(ip, port, netCallback)

	NEIGH_LEN = neigborsLen
	RING = ring.NewRing(OWN_ID, ip, port, 32, NEIGH_LEN)
	if joinIp != "" && joinPort != "" {
		log.Println("Attempting to join: " + ip + ":" + port)
		Join(joinIp, joinPort)

	}
	log.Println("TESTING Put:")
	//testRing(RING)
	//testPut("/Users/karthik/Downloads/key.txt")
	//testGet("key.txt")
	maintenanceLoop(maintenanceTime)
}

func testPut(filePathToUpload string) {
	fileName := strings.Split(filePath, "/")
	hashedFileName := SHAify(fileName[len(fileName)-1])
	putMessage := new(ring.Message).MakePut(hashedFileName, fileEncryptAndSend(filePathToUpload), RING.GetOwner())
	owner := RING.GetOwner()
	peer := RING.Search(hashedFileName).Search(hashedFileName, &owner)
	peer.Connect()
	peer.Send(putMessage.Marshal())
	peer.Close()
}

func testGet(fileName string) {
	hashedFileName := SHAify(fileName)
	getMessage := new(ring.Message).MakeGet(hashedFileName, RING.GetOwner())
	owner := RING.GetOwner()
	peer := RING.Search(hashedFileName).Search(hashedFileName, &owner)
	peer.Connect()
	peer.Send(getMessage.Marshal())
	peer.Close()
}

func testRing(myRing *ring.Ring) {
	mypeer := ring.NewPeer("ddd", "1231", "12212")
	for i := 0; i < 6; i++ {
		myRing.AddNeighbour(*ring.NewPeer("a"+new(big.Int).SetInt64(int64(i)).Text(16), "12", "12322"))
	}

	neighList := myRing.GetNeighbors()
	println("neighLen: ", neighList.Len())
	for i := 0; i < neighList.Len(); i++ {
		log.Println(neighList.Get(i).ToJsonString())
	}

	myRing.AddNeighbour(*mypeer)
	neighList = myRing.GetNeighbors()
	println("neighLen: ", len(neighList))
	for i := 0; i < neighList.Len(); i++ {
		log.Println(neighList.Get(i).ToJsonString() + "\n")
	}

}

func maintenanceLoop(mTime time.Duration) {
	for {
		RING.FixFingers()
		time.Sleep(10 * time.Second) //mTime)
	}

}

func Join(ip, port string) {
	owner := RING.GetOwner()
	entry := ring.NewPeer("", ip, port)
	log.Println("Searching for closest node on ring.")
	closest := entry.Search(owner.ID, &owner) //peerNet.Search(*protocol.NewMessage().MakeSearch(RING.GetOwner().ID, RING.GetOwner()).Marshal(), *entry)
	if closest == nil {
		log.Println("ERROR: could not join network.")
		return
	}
	log.Println("Found: " + closest.ID + ". Attempting notify.")
	neighList := closest.Notify(RING.GetOwner())
	log.Println("Node responded with: ", neighList.Len(), " nodes. Attempting add")
	tmp := append(*neighList, *closest)
	neighList = &tmp

	for i := 0; i < neighList.Len(); i++ {
		RING.AddNeighbour(*neighList.Get(i))
	}
}

func processNotify(message *ring.Message, peer *ring.Peer) {
	neighList := RING.GetNeighbors()
	RING.AddNeighbour(message.Owner)
	res := new(ring.Message).MakeResponse(RING.GetOwner())
	for i := 0; i < neighList.Len(); i++ {
		res.Vars = append(res.Vars, neighList.Get(i).ToJsonString())
	}
	peer.Send(res.Marshal())
	peer.Close()
}

func processSearch(message *ring.Message, peer *ring.Peer) {
	log.Println("Processing search request, Peer: " + peer.ToJsonString())
	term := message.Vars[0]
	println("Searching for: " + term)
	best := RING.Search(term)

	res := new(ring.Message).MakeResponse(RING.GetOwner())
	res.Vars = append(res.Vars, best.ToJsonString())
	peer.Send(res.Marshal())
	peer.Close()
	println("Done processing search")
}

func processPut(message *ring.Message, peer *ring.Peer) {
	log.Println("Node ", peer.ID, "is writing file ", message.Vars[0])
	err := os.WriteFile(message.Vars[0], []byte(message.Vars[1]), 0777)
	if err != nil {
		return
	}
}

func processGet(message *ring.Message, peer *ring.Peer) {
	log.Println("Node", peer.ID, "IP:", peer.Ip, "Port:", peer.Port, " contains requested file")
}

func HandleIncoming(data []byte, conn net.Conn) {
	message := ring.Message{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Println("ERROR UNMARSHALLING NEW CONNECTION")
	}
	message.Owner.SetConn(conn)
	peer := &message.Owner

	switch message.Action {
	case "notify": //Should update neighbor list, and return a list of all known neighbors
		processNotify(&message, peer)
		break
	case "search":
		processSearch(&message, peer)
		break
	case "put":
		processPut(&message, peer)
		break
	case "get":
		processGet(&message, peer)
		break
	case "response":
		break
	case "error":
		break
	}
}

/*

func getOwnerStruct() ring.Peer {
	return ring.Peer{Ip: OWN_IP, Port: OWN_PORT, ID: OWN_ID}
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
			bytes, err := json.Marshal(Message{Owner: getOwnerStruct(), Action: "search", Vars: tmpVars})
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


func fingerSearch(id string) *ring.Peer {
	var foundPeer *ring.Peer = nil
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

func searchResponse(message Message, conn net.Conn) {
	log.Println("Building search response.")
	log.Println("Vars: ", message.Vars, "LEN: ", len(message.Vars))
	//message.Vars = []string(message.Vars)
	var foundPeer ring.Peer

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
	response, err := json.Marshal(Message{
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

func respondToNotify(message Message, conn net.Conn) {

	newList := new(PeerList)

	newList.Append(message.Owner)
	refreshNeighbours(newList)

	var resList []string

	for i := 0; i < predecessors.Len(); i++ {
		preBytes, _ := json.Marshal(predecessors[i])
		resList = append(resList, string(preBytes))
	}

	response := Message{Action: "notifyResponse", Owner: getOwnerStruct(), Vars: resList}

	jsonres, _ := json.Marshal(response)
	sendToPeer(conn, jsonres)

}

func handleIncoming(message Message, conn net.Conn) {
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
		Message{
			Action: "search",
			Owner: ring.Peer{
				Ip:   OWN_IP,
				Port: OWN_PORT,
				ID:   OWN_ID,
			},
			Vars: []string{OWN_ID},
		})
	if err != nil {
		log.Println("ERROR MARSHALLING JOIN MESSAGE: ", err.Error())
	}
	succ := search(message, ring.Peer{Ip: ip, Port: port})
	notify(succ)

	println("FINISHED JOIN")

}


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


func notify(peer ring.Peer) {
	sucConn := connectToPeer(peer.Ip, peer.Port)
	message, err := json.Marshal(
		Message{
			Owner:  getOwnerStruct(),
			Action: "notify"})

	if err != nil {
		log.Println("ERROR MARSHALLING NOTIFY MESSAGE: ", err.Error())
	}
	sendToPeer(sucConn, message)
	//sucConn.Close()
	response := Message{}
	err = json.Unmarshal(listenForData(sucConn), &response)
	rawNeighbors := response.Vars
	var neighbors PeerList
	for i := 0; i < len(rawNeighbors); i++ {
		tmpPeer := ring.Peer{}
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


type PeerList []ring.Peer


func (a PeerList) Len() int {
	return len(a)
}
func (a PeerList) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}
func (a PeerList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a PeerList) Append(b ring.Peer) {
	a = append(a, b)
}
*/
