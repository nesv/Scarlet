// Clustering-related functions, types and interfaces.
//
package main

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	redisSlaveChannel = make(chan *RedisHost)
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

	go func() {
		var slave *RedisHost
		for {
			select {
			case slave = <-redisSlaveChannel:
				Redis.Slaves = append(Redis.Slaves, slave)
				fmt.Printf("REPLICATION\t%s added to list of slaves for %s\n", slave.Addr, Redis.Addr)
			}
		}
	}()

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
		redisSlaveChannel <- slave
	}
	return
}

func startAddSlaveListener() {
	var slave *RedisHost
	for {
		select {
		case slave = <-redisSlaveChannel:
			Redis.Slaves = append(Redis.Slaves, slave)
			fmt.Println("REPLICATION\t%s added to list of slaves for %s", slave.Addr, Redis.Addr)
		}
	}
}
