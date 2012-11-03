// configu.go
//
// Types, and functions relating to Scarlet's configuration.
//
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type ServerBlock struct {
	Enabled       bool   `json:enabled`
	ListenAddress string `json:listenAddress`
	Port          int    `json:port`
}

type RedisBlock struct {
	Protocol        string "protocol"
	Host            string `json:host`
	Port            int    `json:port`
	PropagateWrites bool   `json:propagateWritesToMaster`
	Password        string `json:password`
	DisableInfo     bool   `json:disableInfo`
}

func (r RedisBlock) InfoDisabled() (p bool) {
	p = r.DisableInfo
	return
}

func (r RedisBlock) ConnectAddr() (addr string) {
	addr = fmt.Sprintf("%s:%d", r.Host, r.Port)
	return
}

type Configuration struct {
	HTTP  ServerBlock `json:http`
	TCP   ServerBlock `json:tcp`
	Redis RedisBlock  `json:redis`
}

func LoadConfig(path string) (config *Configuration, err error) {
	var data []byte
	data, err = ioutil.ReadFile(path)

	if *debug {
		println("debug:", "read in", len(data), "bytes")
	}

	if err != nil {
		return
	}

	var c Configuration
	err = json.Unmarshal(data, &c)
	if err != nil {
		return
	}
	config = &c
	return
}

func (c *Configuration) HttpAddress() (addr string) {
	addr = fmt.Sprintf("%s:%d", c.HTTP.ListenAddress, c.HTTP.Port)
	return
}

func (conf *Configuration) Validate() (err error) {
	// Validate the value set for the Redis protocol.
	//
	if conf.Redis.Protocol != "unix" && conf.Redis.Protocol != "tcp" {
		err = errors.New("Redis protocol must be one of \"tcp\" or \"unix\"")
	}
	return
}
