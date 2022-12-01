package main

import "time"

var predecessors [3]Client
var successors [3]Client
var fingerTable [161]Client

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

}
