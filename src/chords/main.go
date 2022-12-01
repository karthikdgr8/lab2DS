package main

import "time"

func main() {
	// parse arguments
	createInstance("127.0.0.1", "12323")
	start()
	print("Sleeping")

	time.Sleep(100 * time.Millisecond)
}
