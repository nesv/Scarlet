// Clustering-related functions, types and interfaces.
//
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Auto-discovers any/all slaves connected to the Redis host we are currently
// connected to, and any masters our Redis host may be a slave of.
//
// This function should definitely be called as a goroutine.
//
func StartAutoDiscovery() {
	conn, err := Database.DB(0)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Auto-discovery halted")
		return
	}
	hostInfo, err := GetHostInfo(conn, "replication")

	// From the info we have from our current connection, let's propagate
	// "downwards" if we are a master.
	//
	switch hostInfo["role"] {
	case "master":
		nslaves, err := strconv.Atoi(hostInfo["connected_slaves"])
		if err != nil {
			fmt.Printf("Error converting %v to an integer\n", hostInfo["connected_slaves"])
			return
		}
		for i := 0; i < nslaves; i++ {
			go ConnectToSlave(hostInfo[fmt.Sprintf("slave%d", i)])
		}
	}
	return
}

// Connects to a Redis slave.
//
func ConnectToSlave(info string) {
	// Form a connection string for the
	slaveInfo := strings.Split(info, ",")
	addr := fmt.Sprintf("%s:%s", slaveInfo[0], slaveInfo[1])
	fmt.Println("REPLICATION\tConnecting to slave", addr)
	return
}
