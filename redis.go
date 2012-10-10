// redis.go
//
// Utility functions for common operations on Redis hosts.
//
package main

import (
	"github.com/simonz05/godis/redis"
	"regexp"
	"strconv"
	"strings"
)

var (
	InfoDbRegex = regexp.MustCompile(`db(\d{1,3})`)
)

type ConnectionMap struct {
	netaddr     string
	password    string
	client      *redis.Client
	connections map[int]*redis.Client
}

func NewConnectionMap(netaddr, password string) (cm *ConnectionMap) {
	cm = &ConnectionMap{netaddr: netaddr, password: password}
	return
}

func (c *ConnectionMap) Add(db int, rc *redis.Client) {
	c.connections[db] = rc
	return
}

func (cm *ConnectionMap) PopulateConnections() (err error) {
	client := redis.New(cm.netaddr, 0, cm.password)
	info, e := GetHostInfo(client)
	if e != nil {
		err = e
		return
	}
	conns := make(map[int]*redis.Client)
	for k, _ := range info {
		if InfoDbRegex.MatchString(k) {
			matches := InfoDbRegex.FindStringSubmatch(k)
			if matches == nil {
				continue
			}
			println("INFO", "Found", matches[0], matches[1])
			dbnum, e := strconv.Atoi(matches[1])
			if e != nil {
				err = e
				return
			}
			conns[dbnum] = redis.New(cm.netaddr, dbnum, cm.password)
		}
	}
	cm.connections = conns
	return
}

func GetHostInfo(c *redis.Client) (info map[string]string, err error) {
	elem, err := c.Info()
	if err != nil {
		return
	}
	items := strings.Split(elem.String(), "\r\n")
	info = make(map[string]string)
	for i := 0; i < len(items); i++ {
		if len(items[i]) == 0 {
			break
		}
		opt := strings.Split(items[i], ":")
		info[opt[0]] = opt[1]
	}
	return
}
