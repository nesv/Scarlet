// main.go
//
// Main source file for Scarlet.
//
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	DefaultListenAddress = ":6380"
	Version              = "0.7.1"
	DefaultRedisAddress  = "127.0.0.1:6379"
)

var (
	ListenAddress   = flag.String("a", DefaultListenAddress, "The address Scarlet should listen on.")
	configPath      = flag.String("c", "scarlet.conf.json", "Specify the configuration file")
	debug           = flag.Bool("d", false, "Enable debugging")
	RedisAddress    = flag.String("r", DefaultRedisAddress, "The upstream Redis host to connect to")
	RedisPassword   = flag.String("rp", "", "Password to use when authenticating to the upstream Redis host")
	ReplicationMode = flag.Bool("replication", false, "Enable replication mode")
	config          *Configuration
	Redis           *RedisHost
	systemSignals   = make(chan os.Signal)
)

func main() {
	fmt.Println("Starting scarlet", Version)
	flag.Parse()
	if *debug {
		fmt.Println("debug:", "debugging enabled")
	}

	// Load the configuration
	//
	if *debug {
		fmt.Println("debug:", "using configuration file", *configPath)
	}
	config, err := LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	if config.Redis.InfoDisabled() {
		fmt.Println("Retrieving node information is disabled")
	}

	// Connect to the initial Redis host
	//
	if *RedisAddress != DefaultRedisAddress {
		Redis, err = ConnectToRedisHost(*RedisAddress, *RedisPassword)
	} else {
		Redis, err = ConnectToRedisHost(config.Redis.ConnectAddr(), config.Redis.Password)
	}

	// Start replication, if it was enabled.
	//
	if *ReplicationMode {
		go StartAutoDiscovery()
	}

	// If the HTTP server was enabled in the configuration, start it.
	//
	if config.HTTP.Enabled {
		if *ListenAddress != DefaultListenAddress {
			go startHttp(*ListenAddress)
		} else {
			go startHttp(config.HttpAddress())
		}
	}

	// Start the mainloop (the signal listener)
	//
	startSignalListener()
	return
}

func startSignalListener() {
	signal.Notify(systemSignals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP)
	for {
		sig := <-systemSignals
		switch sig {
		case syscall.SIGINT:
			fmt.Println("caught SIGINT; exiting")
			os.Exit(0)
		case syscall.SIGKILL:
			fmt.Println("caught SIGKILL; exiting")
			os.Exit(0)
		case syscall.SIGHUP:
			fmt.Println("caught SIHUP; reloading...")
			// call a function to reload the configuration, and
			// re-initialize any connections (if necessary)
		}
	}
}

func Debug(message string) {
	if *debug {
		fmt.Printf("debug\t%s\n", message)
	}
}
