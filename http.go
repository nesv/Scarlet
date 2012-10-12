/* http.go
 *
 * Provides the HTTP interface for Scarlet.
 */
package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	urlRegex         = regexp.MustCompile("^/([0-9]{1,2})(/(.+))?(/(ttl|type))?")
	querystringRegex = regexp.MustCompile(`(\?.*)$`)
)

func startHttp(listenAddr string) {
	// URL-to-handler func mappings
	//
	http.HandleFunc("/info", GetInformation)
	http.HandleFunc("/favicon.ico", Favicon)
	http.HandleFunc("/", DispatchRequest)

	// Start listening for requests
	//
	println("Scarlet HTTP listening on", listenAddr)
	panic(http.ListenAndServe(listenAddr, nil))
	return
}

func GetInformation(rw http.ResponseWriter, req *http.Request) {
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

// Dispatches the incoming request to the proper action handler, depending on 
// the HTTP method that was used.
//
func DispatchRequest(rw http.ResponseWriter, req *http.Request) {
	var response R
	switch req.Method {
	case "GET":
		response = HandleReadOperation(req)

	case "POST":
		response = HandleCreateOperation(req)

	case "PUT":
		response = HandleUpdateOperation(req)
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, response)
	return
}

// Handles HTTP GET requests, which are intended for retrieving data.
//
func HandleReadOperation(req *http.Request) (response R) {
	url := querystringRegex.ReplaceAllString(req.URL.String(), "")
	matches := urlRegex.FindStringSubmatch(url)
	if matches == nil {
		response = R{"result": nil, "error": "No database number specified."}
		return
	}

	// Parse out the database number
	//
	db := matches[1]
	dbnum, err := strconv.Atoi(strings.TrimLeft(db, "/"))
	if err != nil {
		response = R{"result": nil, "error": err}
		return
	}
	if *debug {
		println("debug:", "DB #", dbnum)
	}

	// Get a Redis client for the specified database number.
	//
	client := Database.DB(dbnum)

	// Parse out the key name
	//
	key := matches[3]
	if len(key) == 0 {
		// The length of the key name is zero, so just list all
		// of the keys in the database.
		//
		fmt.Println("KEYS", "*")
		keys, err := client.Keys("*")
		if err != nil {
			response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		} else {
			response = R{"result": keys, "error": nil}
		}
		return
	}

	// Get the key type, so that we know how to properly format the
	// response.
	//
	keyType, err := client.Type(key)
	if err != nil {
		response = R{"result": nil, "error": err}
		return
	}

	// Format the response according to the type the key holds.
	//
	switch keyType {
	case "string":
		println("GET", key)
		v, _ := client.Get(key)
		response = R{"result": v, "error": nil}

	case "set":
		println("SMEMBERS", key)
		v, _ := client.Smembers(key)
		response = R{"result": v.StringArray(), "error": nil}

	case "zset":
		println("ZRANGE", key, 0, -1)
		v, _ := client.Zrange(key, 0, -1)
		response = R{"result": v.StringArray(), "error": nil}

	case "list":
		println("LRANGE", key, 0, -1)
		v, _ := client.Lrange(key, 0, -1)
		response = R{"result": v.StringArray(), "error": nil}

	case "hash":
		if field := req.FormValue("field"); field != "" {
			println("HGET", key, field)
			v, _ := client.Hget(key, field)
			response = R{"result": v.String(), "error": nil}
		} else {
			println("HGETALL", key)
			reply, _ := client.Hgetall(key)
			response = R{"result": reply.StringMap(), "error": nil}
		}

	default:
		e := fmt.Sprintf("Unknown type for key %s: %s", key, keyType)
		response = R{"result": nil, "error": e}
	}
	return
}

// Handles HTTP POST requests, intended for creating new keys.
//
func HandleCreateOperation(req *http.Request) (response R) {
	e := "Create operations have not yet been implemented."
	response = R{"result": nil, "error": e}
	return
}

// Handles HTTP PUT requests, inteded for updating keys.
//
func HandleUpdateOperation(req *http.Request) (response R) {
	e := "Update operations have not yet been implemented."
	response = R{"result": nil, "error": e}
	return
}

func Favicon(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}
