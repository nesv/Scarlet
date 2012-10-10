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
	querystringRegex = regexp.MustCompile(`(\?.*)$`)
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
	config, _ := LoadConfig(*configPath)
	if config.Redis.InfoDisabled() {
		e := "Retrieving node information has been disabled."
		response = R{"result": nil, "error": e}
		fmt.Fprint(rw, response)
		return
	}
	println("INFO")
	info, err := GetHostInfo(redisClient)
	response = R{"result": info, "error": err}
	fmt.Fprint(rw, response)
	return
}

func httpDispatcher(rw http.ResponseWriter, req *http.Request) {
	url := querystringRegex.ReplaceAllString(req.URL.String(), "")
	matches := urlRegex.FindStringSubmatch(url)
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
	if *debug {
		println("Key:", key)
	}

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

	case "hash":
		if field := req.FormValue("field"); field != "" {
			println("HGET", key, field)
			v, _ := redisClient.Hget(key, field)
			response = R{"result": v.String(), "error": nil}
		} else {
			println("HGETALL", key)
			reply, _ := redisClient.Hgetall(key)
			response = R{"result": reply.StringMap(), "error": nil}
		}
		
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
