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
	hostInfo, err := Redis.Info("replication")
	if err != nil {
		fmt.Println("REPLICATION\tError while fetching replication info from server.")
		fmt.Println("REPLICATION\tHalting auto-discovery.")
	}

	var redisSlaveChannel = make(chan *RedisHost)
	go func() {
		for {
			select {
			case slave, openp := <-redisSlaveChannel:
				if openp {
					Redis.Slaves = append(Redis.Slaves, slave)
					fmt.Printf("REPLICATION\t%s added to list of slaves for %s\n", slave.Addr, Redis.Addr)
				} else {
					Debug("Closed slave channel; terminating inline gofunc")
					break
				}
			}
		}
	}()

	// From the info we have from our current connection, let's propagate
	// "downwards" by connecting to any slaves.
	//
	nslaves, err := strconv.Atoi(hostInfo["connected_slaves"])
	if err != nil {
		fmt.Printf("Error converting %v to an integer\n", hostInfo["connected_slaves"])
		return
	}
	for i := 0; i < nslaves; i++ {
		go ConnectToSlave(hostInfo[fmt.Sprintf("slave%d", i)], redisSlaveChannel)
	}

	// If our Redis host's role is a slave, then let's establish a connection
	// to the master.
	//
	if hostInfo["role"] == "slave" {
		masterAddress := fmt.Sprintf("%s:%s", hostInfo["master_host"], hostInfo["master_port"])
		fmt.Println("REPLICATION\tConnecting to master", masterAddress)
		if master, err := ConnectToRedisHost(masterAddress, ""); err != nil {
			fmt.Println("REPLICATION\tError: Could not connect to master")
		} else {
			fmt.Println("REPLICATION\tConnected to master")
			Redis.Master = master
		}
	}
	return
}

// Connects to a Redis slave.
//
func ConnectToSlave(info string, rsch chan *RedisHost) {
	slaveInfo := strings.Split(info, ",")
	addr := fmt.Sprintf("%s:%s", slaveInfo[0], slaveInfo[1])
	fmt.Println("REPLICATION\tConnecting to slave", addr)

	// NOTE
	// This is making the gross assumption that the slave we are connecting to
	// does not require a password.
	//
	if slave, err := ConnectToRedisHost(addr, ""); err != nil {
		fmt.Println("REPLICATION\tError:", err)
		return
	} else {
		rsch <- slave
	}
	return
}
