package main

import (
	"lab1DS/src/control"
	"lab1DS/src/sec"
	"log"
	"math/big"
	"math/rand"
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

	ip := "0.0.0.0" // defaults to localhost.
	port := "12323" //default port
	maintenanceTime := 15000 * time.Millisecond
	neighLen := 3
	joinIp := ""
	joinPort := ""

	for i := 1; i < len(os.Args)-1; i++ {

		switch os.Args[i] {
		case "-a":
			if os.Args[i+1] == "empty" {
				break
			}
			ip = os.Args[i+1]
			log.Println("Binding to IP:", ip)
			break
		case "-p":
			if os.Args[i+1] == "empty" {
				break
			}
			port = os.Args[i+1]
			log.Println("Binding Port:", port)
			break
		case "--ja": //Join ip
			if os.Args[i+1] == "empty" {
				break
			}
			joinIp = os.Args[i+1]
			log.Println("Joining IP:", joinIp)
			break
		case "--jp": //join port
			if os.Args[i+1] == "empty" {
				break
			}
			joinPort = os.Args[i+1]
			log.Println("Joining Port:", joinPort)
			break
		case "--ts":
			if os.Args[i+1] == "empty" {
				break
			}
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			log.Println("Stabilizing Time: ", os.Args[i+1])
			break
		case "--tff":
			if os.Args[i+1] == "empty" {
				break
			}
			log.Println("Maintenance Time: ", os.Args[i+1])
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			break
		case "--tcp":
			if os.Args[i+1] == "empty" {
				break
			}
			maintInt, _ := strconv.Atoi(os.Args[i+1])
			maintenanceTime = time.Duration(maintInt) * time.Millisecond
			log.Println("Check Predecessor Time:" + os.Args[i+1])
			break
		case "-i":
			if os.Args[i+1] == "empty" {
				break
			}
			OWN_ID = os.Args[i+1]
			println("Overridden ID: ", OWN_ID)
			break
		case "-r":
			if os.Args[i+1] == "empty" {
				break
			}
			neighLen, _ = strconv.Atoi(os.Args[i+1])
			println("Neigh Len: ", neighLen)
		default:
			continue
		}
	}

	if OWN_ID == "" {
		rand.Seed(time.Now().UnixNano())
		OWN_ID = sec.SHAify(new(big.Int).SetInt64(int64(rand.Intn(500000))).Text(16))
		log.Println("Setting random ID: ", OWN_ID)
	}

	log.Println("---------------------------STARTING PEER---------------------------")
	control.StartUp(ip, port, neighLen, maintenanceTime, OWN_ID, joinIp, joinPort) //TODO: Make neighbours len dynamic.

}
