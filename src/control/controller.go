package control

import (
	"bufio"
	"lab1DS/src/becauseGO"
	"lab1DS/src/peerNet"
	"lab1DS/src/ring"
	"lab1DS/src/sec"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var OWN_ID string
var RING *ring.Ring
var NEIGH_LEN int

/*
This function is the main entry point for the program, where it sets up the network and ring structures, starts the
maintenance loop and console prompt, and joins the ring if necessary.
*/
func StartUp(ip, port string, neigborsLen int, maintenanceTime time.Duration, ownID, joinIp, joinPort string) {
	OWN_ID = ownID
	netCallback := becauseGO.Callback{Callback: HandleIncoming}
	sec.GetKeyFromFile()                                 // Read encryption key from file to memory
	peerNet.CreateNetworkInstance(ip, port, netCallback) // Create a network with IP and Port

	NEIGH_LEN = neigborsLen
	RING = ring.NewRing(OWN_ID, ip, port, 32, NEIGH_LEN) // Create a new ring
	if joinIp != "" && joinPort != "" {
		log.Println("Attempting to join: " + joinIp + ":" + joinPort)
		Join(joinIp, joinPort) // If join IP and Port specified, try joining and exchange peerLists

	}

	go promptCmd()
	maintenanceLoop(maintenanceTime)
}

/*
This function implements a distributed file sharing system, where users can store and retrieve files from a network of peers.
The makeGet and makePut functions are likely responsible for initiating the file transfer process by sending requeststo other peers in the network.
*/
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

	// Create a new message with the name of the file to be pushed to network and the encrypted file itself
	putMessage := new(ring.Message).MakePut(hashedFileName, sec.GetEncryptedFile(filePathToUpload), RING.GetOwner())

	owner := RING.GetOwner()

	// Search for the position of the file in the ring i.e the peer that has to hold the file
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
					log.Println("ERROR SENDING FILE")
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

/*
This function is part of a distributed file-sharing system that uses a ring topology.
The function is used to initiate a file retrieval operation by searching for the node responsible for storing the file
and sending a request to that node to retrieve the file.
The function first computes the hash of the filename, identifies the closest known neighbor to the hashed
filename, and searches for the successor node in the ring.
Once it finds the successor node, it sends a GET request to retrieve the file and waits for the response.
*/
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

/*
This function maintains a logical ring topology, where each node periodically performs maintenance tasks to update
the topology and detect failed nodes. This function is used to ensure that the ring topology remains stable and all
nodes are aware of their neighbors.
*/
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

/*
The function is used to join an existing ring network by connecting to a known node, obtaining a neighbor
list from the node, and updating the ring topology with the new node and its neighbors.
*/
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

/*
This function respond to a notification message from a neighbor by returning a list of all the neighbors in the ring, adding the peer
to the neighbor list, and then closing the connection, where each node maintains information
about its neighbors in the ring, and nodes can join or leave the ring dynamically.
*/
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

/*
The function is used to process a search request from a peer by finding the closest known node to the search
term and returning its JSON representation in the response message.
This function may be used to implement a distributed search mechanism where peers can search for files or other
resources in the network by sending search requests to the closest known nodes.
*/
func processSearch(message *ring.Message, peer *ring.Peer) {
	term := message.Vars[0]

	best := RING.ClosestKnown(term)
	res := new(ring.Message).MakeResponse(RING.GetOwner())
	res.Vars = append(res.Vars, best.ToJsonString())
	peer.Send(res.Marshal())
	peer.Close()
}

/*
The function is used to handle a "PUT" request from a peer, which contains the filename and contents of the file to be written to disk.
The function writes the file to disk and logs a message indicating that the write was successful.
*/
func processPut(message *ring.Message, peer *ring.Peer) {
	log.Println("Node ", RING.GetOwner().ID, "is writing file ", message.Vars[0])
	err := os.WriteFile(filePath+message.Vars[0], []byte(message.Vars[1]), 0777)
	if err != nil {
		log.Println("ERROR WRITING FILE: ", err)
		return
	}
}

/*
The function allows nodes to share files with each other and is used to handle a "GET" request from a peer, which contains the filename of the file to be retrieved.
The function reads the file from disk, logs a message indicating success or failure, and sends a response to the peer indicating the status of the file retrieval.
*/
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

/*
The function is used to handle incoming network connections from peers, read incoming messages, and dispatch
them to the appropriate processing functions based on the message action.
*/
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
			}
		}

	}

}
