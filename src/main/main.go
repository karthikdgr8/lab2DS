package main

import (
	"lab1DS/src/control"
	"log"
	"os"
	"strconv"
	"time"
)

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
	OWN_ID := ""

	ip := "127.0.0.1" // defaults to localhost.
	port := "12323"   //default port
	maintenanceTime := 30000 * time.Millisecond
	neighLen := 3
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
		case "-r":
			neighLen, _ = strconv.Atoi(os.Args[i+1])
		default:
			continue
		}
	}

	//testSetup()

	if OWN_ID == "" {
		log.Println("Setting random ID.")
		OWN_ID = "172b17"

	}

	log.Println("---------------------------STARTING PEER---------------------------")
	control.StartUp(ip, port, neighLen, maintenanceTime, OWN_ID, joinIp, joinPort) //TODO: Make neighbours len dynamic.

}
