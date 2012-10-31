// redis.go
//
// Utility functions for common operations on Redis hosts.
//
package main

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"regexp"
	"strconv"
	"strings"
)

var (
	InfoDbRegex = regexp.MustCompile(`db(\d{1,3})`)
)

// An idiomatic function to create a new connection to a Redis host, and
// subsequently authenticate, and select a database.
//
func ConnectToRedisHost(addr, password string, db int) (c redis.Conn, err error) {
	conn, e := redis.Dial("tcp", addr)
	if e != nil {
		err = e
		return
	}

	// Did the user specify a password?
	//
	if len(password) > 0 {
		ok, e := redis.Bool(conn.Do("AUTH", password))
		if e != nil {
			err = e
			return
		} else if !ok {
			err = errors.New("AUTH failed")
			return
		}
	}

	// Now, change over to the specified database.
	//
	if ok, e := redis.Bool(conn.Do("SELECT", db)); err != nil {
		err = e
		return
	} else if !ok {
		msg := fmt.Sprintf("Could not SELECT database #%d", db)
		err = errors.New(msg)
		return
	}

	c = conn
	return
}

// A ConnectionMap holds a collection of Redis clients, and provides a nice,
// easy way to quickly establish a new connection to a database on a Redis
// host, and get the appropriate connection for the incoming HTTP request.
//
type ConnectionMap struct {
	netaddr     string
	password    string
	client      redis.Conn
	connections map[int]redis.Conn
}

// Creates (and returns) a pointer to a ConnectionMap.
//
func NewConnectionMap(netaddr, password string) (cm *ConnectionMap) {
	cm = &ConnectionMap{netaddr: netaddr, password: password}
	return
}

// Associates a Redis client to the database it is connected to, so it can be
// kept around for future use.
//
func (c *ConnectionMap) Add(db int, rc redis.Conn) {
	c.connections[db] = rc
	return
}

// Returns the appropriate Redis client for the database number provided. If
// there is no client for that database number, then this function will create
// it, and return it.
//
func (c *ConnectionMap) DB(db int) (r redis.Conn, err error) {
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
	r, e := ConnectToRedisHost(c.netaddr, c.password, db)
	if e != nil {
		err = e
		return
	}
	c.Add(db, r)
	return
}

// Populates connections to any database. on the Redis host the ConnectionMap
// was initialized with, that holds data.
//
func (cm *ConnectionMap) PopulateConnections() (err error) {
	client, e := ConnectToRedisHost(cm.netaddr, cm.password, 0)
	if e != nil {
		err = e
		return
	}

	info, e := GetHostInfo(client)
	if e != nil {
		err = e
		return
	}

	conns := make(map[int]redis.Conn)
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
			conn, e := ConnectToRedisHost(cm.netaddr, cm.password, dbnum)
			if e != nil {
				err = e
				return
			}
			conns[dbnum] = conn
		}
	}
	cm.connections = conns
	return
}

// Runs the INFO command on the remote Redis host, and nicely maps the response
// from the server to a string-string map.
//
func GetHostInfo(c redis.Conn) (info map[string]string, err error) {
	v, err := redis.String(c.Do("INFO"))
	items := strings.Split(v, "\r\n")
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
