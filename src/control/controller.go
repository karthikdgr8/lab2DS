package control

import (
	"bufio"
	"lab1DS/src/becauseGO"
	"lab1DS/src/peerNet"
	"lab1DS/src/ring"
	"lab1DS/src/sec"
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
	sec.GetKeyFromFile()
	peerNet.CreateNetworkInstance(ip, port, netCallback)

	NEIGH_LEN = neigborsLen
	RING = ring.NewRing(OWN_ID, ip, port, 32, NEIGH_LEN)
	if joinIp != "" && joinPort != "" {
		log.Println("Attempting to join: " + joinIp + ":" + joinPort)
		Join(joinIp, joinPort)

	}

	go promptCmd()
	maintenanceLoop(maintenanceTime)
}

func promptCmd() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()
		if strings.Contains(text, "Lookup") {
			split := strings.Split(text, " ")
			fileName := split[len(split)-1]
			log.Println("Looking up", fileName)
			makeGet(fileName)
		} else if strings.Contains(text, "StoreFile") {
			split := strings.Split(text, " ")
			tempFilePath := split[len(split)-1]
			split = strings.Split(tempFilePath, "/")
			fileName := split[len(split)-1]
			log.Println("Storing file", fileName)
			makePut(tempFilePath, fileName)
		}
	}
}

/*
Function for searching for finding the relevant peer(s) on the network, and pushing a local file to the peer(s)
*/
func makePut(filePathToUpload, fileName string) {

	hashedFileName := sec.SHAify(fileName)
	log.Println("PUT STRING: ", fileName, "hashes to: ", hashedFileName)
	putMessage := new(ring.Message).MakePut(hashedFileName, sec.GetEncryptedFile(filePathToUpload), RING.GetOwner())
	owner := RING.GetOwner()
	succ := RING.ClosestKnown(hashedFileName).Search(hashedFileName, &owner)

	for i := 0; i < 3; i++ { //Redundancy
		if succ != nil {
			id := succ.ID
			if id != RING.GetOwner().ID {
				succ.Connect()
				log.Println("Storing file on node: ", id)
				log.Println("Node address: ", succ.Ip, ":", succ.Port)
				err := succ.Send(putMessage.Marshal())
				if err == nil {
					println("ERROR SENDING FILE")
				}
				succ.Close()
			}
			succ = succ.Search(succ.ID, &owner)
			if id == succ.ID {
				log.Println("Reached end of network before redundancy requirement met.")
				return
			}
		}
	}
}

func makeGet(fileName string) {
	hashedFileName := sec.SHAify(fileName)
	getMessage := new(ring.Message).MakeGet(hashedFileName, RING.GetOwner())
	owner := RING.GetOwner()
	peer := RING.ClosestKnown(hashedFileName).Search(hashedFileName, &owner)
	succ := peer.Search(hashedFileName, &owner)
	if succ.ID == OWN_ID {
		succ = peer.Search(OWN_ID, &owner)
	}
	succ.Connect()
	succ.Send(getMessage.Marshal())
	res := succ.ReadMessage()
	log.Println("Response from peer: " + res.Vars[0])
	succ.Close()
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
		log.Println("Maintaining network.")
		RING.Stabilize()
		RING.FixFingers()
		list := RING.GetNeighbors()
		log.Print("Known neighbours after maintenance: ")
		for i := 0; i < list.Len(); i++ {
			print(list.Get(i).ID, ", ")
		}
		println()
		time.Sleep(mTime)
	}

}

func Join(ip, port string) {
	owner := RING.GetOwner()
	entry := ring.NewPeer("", ip, port)
	log.Println("Searching for closest node on ring.")
	log.Println("Entrypoint: ", ip, ":", port)
	closest := entry.Search(owner.ID, &owner)
	if closest == nil {
		log.Println("ERROR: could not join network.")
		return
	}
	log.Println("Found: " + closest.ID + ". Attempting notify.")
	neighList := closest.Notify(RING.GetOwner())
	println("Closest found node ID: ", closest.ID)
	log.Println("Node responded with: ", neighList.Len(), " nodes. Attempting add.")
	RING.AddNeighbour(*closest)
	for i := 0; i < neighList.Len(); i++ {
		if neighList.Get(i).ID != closest.ID && neighList.Get(i).ID != RING.GetOwner().ID {
			if neighList.Get(i).Notify(owner) != nil {
				RING.AddNeighbour(*neighList.Get(i))
			}
		}
	}

	log.Print("Join successful")
}

func processNotify(message *ring.Message, peer *ring.Peer) {
	neighList := RING.GetNeighbors()

	res := new(ring.Message).MakeResponse(RING.GetOwner())
	for i := 0; i < neighList.Len(); i++ {
		res.Vars = append(res.Vars, neighList.Get(i).ToJsonString())
	}
	RING.AddNeighbour(*peer)
	peer.Send(res.Marshal())
	peer.Close()

}

func processSearch(message *ring.Message, peer *ring.Peer) {
	term := message.Vars[0]

	best := RING.ClosestKnown(term)
	res := new(ring.Message).MakeResponse(RING.GetOwner())
	res.Vars = append(res.Vars, best.ToJsonString())
	peer.Send(res.Marshal())
	peer.Close()
}

func processPut(message *ring.Message, peer *ring.Peer) {
	log.Println("Node ", RING.GetOwner().ID, "is writing file ", message.Vars[0])
	err := os.WriteFile(filePath+message.Vars[0], []byte(message.Vars[1]), 0777)
	if err != nil {
		log.Println("ERROR WRITING FILE: ", err)
		return
	}
}

func processGet(message *ring.Message, peer *ring.Peer) {
	file, _ := os.ReadFile(filePath + message.Vars[0])
	var resString string
	if file != nil {
		log.Println("Node", RING.GetOwner().ID, " contains requested file")
		resString = "Node " + RING.GetOwner().ID + " contains requested file"
	} else {
		log.Println("Node", RING.GetOwner().ID, " does not contain requested file")
		resString = "Node " + RING.GetOwner().ID + " could not find file"
	}
	res := new(ring.Message).MakeResponse(RING.GetOwner())
	res.Vars = append(res.Vars, resString)
	peer.Send(res.Marshal())
	peer.Close()
}

func HandleIncoming(conn net.Conn) {

	peer := ring.FromNetwork(conn)
	if peer != nil {
		message := peer.ReadMessage()
		if message != nil {
			peer.Port = message.Owner.Port
			peer.ID = message.Owner.ID
			switch message.Action {
			case "notify": //Should update neighbor list, and return a list of all known neighbors
				processNotify(message, peer)
				break
			case "search":
				processSearch(message, peer)
				break
			case "put":
				processPut(message, peer)
				break
			case "get":
				processGet(message, peer)
				break
			case "response":
				break
			case "error":
				break
			}
		}

	}

}
