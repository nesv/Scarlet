/* http.go
 *
 * Provides the HTTP interface for Scarlet.
 */
package main

import (
	"net/http"
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

var (
	urlRegex = regexp.MustCompile("^/([0-9]{1,2})(/(.+))?(/(ttl|type))?")
)

func startHttp(listenAddr string) {
	// URL-to-handler func mappings
	//
	http.HandleFunc("/info", httpGetInfo)
	http.HandleFunc("/favicon.ico", httpFavicon)
	http.HandleFunc("/", httpDispatcher)

	// Start listening for requests
	//
	println("Scarlet HTTP listening on", listenAddr)
	panic(http.ListenAndServe(listenAddr, nil))
	return
}

func httpGetInfo(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	var response R
	config, _ := loadConfig(*configPath)
	if config.Redis.InfoDisabled() {
		e := "Retrieving node information has been disabled."
		response = R{"result": nil, "error": e}
		fmt.Fprint(rw, response)
		return
	}
	println("INFO")
	elem, err := redisClient.Info()
	if err != nil {
		response = R{"result": nil, "error": err}
		fmt.Fprint(rw, response)
	}
	items := strings.Split(elem.String(), "\r\n")
	info := make(map[string]string)
	for i := 0; i < len(items); i++ {
		if len(items[i]) == 0 {
			break
		}
		opt := strings.Split(items[i], ":")
		if *debug {
			println("debug:", opt[0])
		}
		if len(opt) > 2 {
			e := fmt.Sprintf("Info item %s has more than two fields", opt[0])
			response = R{"result": nil, "error": e}
			fmt.Fprint(rw, response)
			return
		}
		info[opt[0]] = opt[1]
	}
	response = R{"result": info, "error": nil}
	fmt.Fprint(rw, response)
	return
}

func httpDispatcher(rw http.ResponseWriter, req *http.Request) {
	matches := urlRegex.FindStringSubmatch(req.URL.String())
	if matches == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, ":(")
		return
	}

	// Parse out the database number
	//
	db := matches[1]
	dbnum, err := strconv.Atoi(strings.TrimLeft(db, "/"))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, err)
		return
	}
	if *debug {
		println("debug:", "DB #", dbnum)
	}

	// Parse out the key name
	//
	key := matches[3]
	println("Key:", key)

	// Initialize a variable of type R, to hold the response that will
	// eventually be JSON encoded.
	//
	var response R
	rw.Header().Set("Content-Type", "application/json")

	if len(key) == 0 {
		if req.Method == "GET" {
			// The length of the key name is zero, so just list all
			// of the keys in the database.
			//
			println("KEYS", "*")
			keys, err := redisClient.Keys("*")
			if err != nil {
				response = R{"result": nil, "error": err}
			} else {
				response = R{"result": keys, "error": nil}
			}
			fmt.Fprint(rw, response)
			return
		} else {
			rw.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// Get the key type, so that we know how to properly format the
	// response.
	//
	keyType, err := redisClient.Type(key)
	if err != nil {
		response = R{"result": nil, "error": err}
		fmt.Fprint(rw, response)
		return
	}
	
	// Format the response according to the type the key holds.
	//
	switch keyType {
	case "string":
		println("GET", key)
		v, _ := redisClient.Get(key)
		response = R{"result": v, "error": nil}
		
	case "set":
		println("SMEMBERS", key)
		v, _ := redisClient.Smembers(key)
		response = R{"result": v.StringArray(), "error": nil}
		
	case "zset":
		println("ZRANGE", key, 0, -1)
		v, _ := redisClient.Zrange(key, 0, -1)
		response = R{"result": v.StringArray(), "error": nil}

	case "list":
		println("LRANGE", key, 0, -1)
		v, _ := redisClient.Lrange(key, 0, -1)
		response = R{"result": v.StringArray(), "error": nil}
		
	default:
		e := fmt.Sprintf("Unknown type for key %s: %s", key, keyType)
		response = R{"result": nil, "error": e}
	}
	
	fmt.Fprint(rw, response)
	return
}

func httpFavicon(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}
