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
	Version              = "0.7.0"
	DefaultRedisAddress  = "127.0.0.1:6379"
)

var (
	ListenAddress = flag.String("a", DefaultListenAddress, "The address Scarlet should listen on.")
	configPath    = flag.String("c", "scarlet.conf.json", "Specify the configuration file")
	debug         = flag.Bool("d", false, "Enable debugging")
	RedisAddress  = flag.String("r", DefaultRedisAddress, "The upstream Redis host to connect to")
	RedisPassword = flag.String("rp", "", "Password to use when authenticating to the upstream Redis host")
	config        *Configuration
	Database      *ConnectionMap
	systemSignals = make(chan os.Signal)
)

func main() {
	println("Starting scarlet", Version)
	flag.Parse()
	if *debug {
		println("debug:", "debugging enabled")
	}

	// Load the configuration
	//
	if *debug {
		println("debug:", "using configuration file", *configPath)
	}
	config, err := LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	if config.Redis.InfoDisabled() {
		println("Retrieving node information is disabled")
	}

	// Connect to the initial Redis host
	//
	if *RedisAddress != DefaultRedisAddress {
		Database = NewConnectionMap(*RedisAddress, *RedisPassword)
	} else {
		Database = NewConnectionMap(config.Redis.ConnectAddr(), config.Redis.Password)
	}
	err = Database.PopulateConnections()
	if err != nil {
		fmt.Printf("FATAL\tCould not populate connections: %s\n", err)
		return
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
			println("caught SIGINT; exiting")
			os.Exit(0)
		case syscall.SIGKILL:
			println("caught SIGKILL; exiting")
			os.Exit(0)
		case syscall.SIGHUP:
			println("caught SIHUP; reloading...")
			// call a function to reload the configuration, and
			// re-initialize any connections (if necessary)
		}
	}
}
