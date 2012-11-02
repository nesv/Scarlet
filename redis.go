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

// A ConnectionMap holds a collection of Redis clients, and provides a nice,
// easy way to quickly establish a new connection to a database on a Redis
// host, and get the appropriate connection for the incoming HTTP request.
//
type ConnectionMap struct {
	netaddr     string
	password    string
	client      *redis.Client
	connections map[int]*redis.Client
}

// Creates (and returns) a pointer to a ConnectionMap.
//
func NewConnectionMap(netaddr, password string) (cm *ConnectionMap) {
	cm = &ConnectionMap{netaddr: netaddr, password: password}
	return
}

// Returns a list of database numbers for which there are currently connections
// established.
//
func (c *ConnectionMap) NConnections() (dbs []int) {
	for k, _ := range c.connections {
		dbs = append(dbs, k)
	}
	return
}

// Associates a Redis client to the database it is connected to, so it can be
// kept around for future use.
//
func (c *ConnectionMap) Add(db int, rc *redis.Client) {
	c.connections[db] = rc
	return
}

// Returns the appropriate Redis client for the database number provided. If
// there is no client for that database number, then this function will create
// it, and return it.
//
func (c *ConnectionMap) DB(db int) (r *redis.Client) {
	client, existsp := c.connections[db]
	if existsp {
		// Yay, we already have a client established to that database!
		//
		r = client
		return
	}

	// Urg, it looks like this is the first time anything has been requested
	// regarding this database. Let's establish a new connection to it, and
	// save it for later.
	//
	// ...then return it (because we're nice like that).
	//
	if *debug {
		println("DEBUG", "Creating new Redis connection to DB #", db)
	}
	client = redis.New(c.netaddr, db, c.password)
	c.Add(db, client)
	r = client
	return
}

// Populates connections to any database. on the Redis host the ConnectionMap
// was initialized with, that holds data.
//
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

// Runs the INFO command on the remote Redis host, and nicely maps the response
// from the server to a string-string map.
//
func GetHostInfo(c *redis.Client) (info map[string]string, err error) {
	elem, err := c.Info()
	if err != nil {
		return
	}
	items := strings.Split(elem.String(), "\r\n")
	info = make(map[string]string)
	for i := 0; i < len(items); i++ {
		if len(items[i]) == 0 || string(items[i][0]) == "#" {
			continue
		}
		opt := strings.Split(items[i], ":")
		info[opt[0]] = opt[1]
	}
	return
}
