// main.go
//
// Main source file for Scarlet.
//
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"github.com/simonz05/godis/redis"
)

var (
	configPath = flag.String("c", "scarlet.conf.json", "Specify the configuration file")
	debug = flag.Bool("d", false, "Enable debugging")
	config Configuration
	redisClient *redis.Client
	systemSignals = make(chan os.Signal)
)

func main() {
	println("Starting scarlet")
	flag.Parse()
	if *debug {
		println("debug:", "debugging enabled")
	}
	
	// Load the configuration
	//
	if *debug {
		println("debug:", "using configuration file", *configPath)
	}
	config, err := loadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	// Connect to the initial Redis host
	//
	redisClient = redis.New(config.Redis.ConnectAddr(), 0, config.Redis.Password)

	// If the HTTP server was enabled in the configuration, start it.
	//
	if config.HTTP.Enabled {
		go startHttp(config.HttpAddress())
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
