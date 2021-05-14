package main

//
// start the coordinator process, which is implemented
// in ../mr/coordinator.go
//
// go run mrcoordinator.go pg*.txt
//
// Please do not change this file.
//

import (
	"fmt"
	"os"
	"time"

	"6.824/mr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrcoordinator inputfiles...\n")
		os.Exit(1)
	}

	startTime := time.Now().UnixNano()

	m := mr.MakeCoordinator(os.Args[1:], 10)

	// YifanLu here
	// build ssh connection to workers and send command "go run -race mrworker.go wc"

	hosts := []string{"192.168.0.132", "192.168.0.184", "192.168.0.33", "192.168.0.199"}
	command := "go run -race mrworker.go wc"
	mr.AwakenWorkers("root", "Ydhlw123", hosts, command)

	for m.Done() == false {
		time.Sleep(time.Second)
	}
	endTime := time.Now().UnixNano()
	seconds := float64((endTime - startTime) / 1e9)
	fmt.Printf("EXECUTION FINISHED!\nRUNNING TIME:%v\n", seconds)

	time.Sleep(time.Second * 10086)
}
