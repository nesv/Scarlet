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
func ConnectToRedisHost(addr, password string, db interface{}) (c redis.Conn, err error) {
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
	if _, e := redis.String(conn.Do("SELECT", db)); e != nil {
		err = e
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

	info, e := GetHostInfo(client, "default")
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
			println("Found", matches[0], matches[1])
			conn, e := ConnectToRedisHost(cm.netaddr, cm.password, matches[1])
			if e != nil {
				err = e
				return
			}
			dbnum, _ := strconv.Atoi(matches[1])
			conns[dbnum] = conn
		}
	}
	cm.connections = conns

	// If replication mode is enabled, let's call our lovely goroutine for
	// auto-discovering masters & slaves.
	//
	if *ReplicationMode {
		go StartAutoDiscovery()
	}

	return
}

// Runs the INFO command on the remote Redis host, and nicely maps the response
// from the server to a string-string map.
//
func GetHostInfo(c redis.Conn, section string) (info map[string]string, err error) {
	switch section {
	case "server", "clients", "memory", "persistence", "stats", "replication":
		fallthrough
	case "cpu", "commandstats", "cluster", "keyspace", "all", "default":
	case "":
		section = "default"
	default:
		err = errors.New(fmt.Sprintf("Invalid INFO section: %s", section))
		return
	}
	v, err := redis.String(c.Do("INFO"))
	items := strings.Split(v, "\r\n")
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

// A type to represent a Redis host.
//
type RedisHost struct {
	Addr      string
	Port      int
	password  string
	Slaves    []RedisSlave
	databases map[int]redis.Conn
}

func (h *RedisHost) Info(section string) (info map[string]string, err error) {
	switch section {
	case "server", "clients", "memory", "persistence", "stats", "replication":
		fallthrough
	case "cpu", "commandstats", "cluster", "keyspace", "all", "default":
	case "":
		section = "default"
	default:
		err = errors.New(fmt.Sprintf("Invalid INFO section: %s", section))
		return
	}
	v, err := redis.String(c.Do("INFO"))
	items := strings.Split(v, "\r\n")
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

func (h *RedisHost) addDatabaseConnection(dbnum int, conn redis.Conn) {
	h.databases[dbnum] = conn
	return
}

func (h *RedisHost) Db(n int) (r redis.Conn) {
	r, existsp := h.databases[n]
	if existsp {
		return
	}
	r, e := ConnectToRedisHost(h.Addr, h.password, n)
	if e != nil {
		fmt.Println("Error:", e)
		return
	}
	h.addDatabaseConnection(db, r)
	return
}
