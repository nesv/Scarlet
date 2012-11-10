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
func ConnectToRedisHost(addr, password string) (r *RedisHost, err error) {
	r = &RedisHost{Addr: addr, password: password}
	err = r.Connect()
	return
}

// A type to represent a Redis host.
//
type RedisHost struct {
	Addr      string
	password  string
	Slaves    []*RedisHost
	databases map[int]redis.Conn
	Master    *RedisHost
}

// Returns the number of currently-active Redis connections for the host we
// were configured to connect to.
func (h *RedisHost) NConnections() (n int) {
	n = len(h.databases)
	return
}

// Connect to the Redis host, and authenticate if necessary.
func (h *RedisHost) Connect() (err error) {
	conn, e := redis.Dial("tcp", h.Addr)
	if e != nil {
		err = e
		return
	}
	if len(h.password) > 0 {
		ok, e := redis.Bool(conn.Do("AUTH", h.password))
		if e != nil {
			err = e
			return
		} else if !ok {
			err = errors.New("AUTH failed")
			return
		}
	}
	if _, e := redis.String(conn.Do("SELECT", 0)); e != nil {
		err = e
	}

	// Initialize the int->redis.Conn map for storing database connections.
	//
	h.databases = make(map[int]redis.Conn)

	// Put our connection in (for database #0) and populate connections.
	//
	h.databases[0] = conn
	h.populateDatabaseConnections()
	return
}

// Create connections to any other databases on this host.
func (h *RedisHost) populateDatabaseConnections() {
	info, err := h.Info("default")
	if err != nil {
		return
	}
	for k, _ := range info {
		if InfoDbRegex.MatchString(k) {
			matches := InfoDbRegex.FindStringSubmatch(k)
			if matches == nil {
				continue
			}
			dbnum, _ := strconv.Atoi(matches[1])
			fmt.Println("Found", matches[0], matches[1])
			c, err := h.connectToDatabase(dbnum)
			if err != nil {
				return
			}
			h.databases[dbnum] = c
		}
	}
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
	c := h.Db(0)
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

// Creates a new connection to a database "db" on *RedisHost "h".
func (h *RedisHost) connectToDatabase(db int) (r redis.Conn, err error) {
	conn, err := redis.Dial("tcp", h.Addr)
	if err != nil {
		return
	}

	if len(h.password) > 0 {
		authenticated, e := redis.Bool(conn.Do("AUTH", h.password))
		if e != nil {
			err = e
			return
		} else if !authenticated {
			err = errors.New("AUTH failed.")
		}
	}

	if _, err = redis.String(conn.Do("SELECT", db)); err != nil {
		return
	}

	fmt.Println("Connected to database #", db)
	r = conn
	return
}

// Get a connection to a database "n".
func (h *RedisHost) Db(n int) (r redis.Conn) {
	r, existsp := h.databases[n]
	if existsp {
		return
	}

	r, e := h.connectToDatabase(n)
	if e != nil {
		fmt.Println("Error:", e)
		return
	}
	h.databases[n] = r
	return
}
