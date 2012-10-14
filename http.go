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
	"errors"
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

type RequestInfo struct {
	DbNum int
	Key string
}

func GetRequestInfo(r *http.Request) (ri *RequestInfo, err error) {
	url := querystringRegex.ReplaceAllString(r.URL.String(), "")
	m := urlRegex.FindStringSubmatch(url)
	if m == nil {
		err = errors.New("Malformed URL")
		return
	}
	dbnum, e := strconv.Atoi(strings.TrimLeft(m[1], "/"))
	if err != nil {
		err = e
		return
	}
	ri = &RequestInfo{DbNum: dbnum, Key: m[3]}
	return
}

// Dispatches the incoming request to the proper action handler, depending on 
// the HTTP method that was used.
//
func DispatchRequest(rw http.ResponseWriter, req *http.Request) {
	var response R
	if info, err := GetRequestInfo(req); err == nil {
		switch req.Method {
		case "GET":
			response = HandleReadOperation(req, info)

		case "POST":
			response = HandleCreateOperation(req, info)

		case "PUT":
			response = HandleUpdateOperation(req, info)
		}
	} else {
		response = R{"result": nil, "error": err}
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, response)
	return
}

// Handles HTTP GET requests, which are intended for retrieving data.
//
func HandleReadOperation(req *http.Request, info *RequestInfo) (response R) {
	// Get a Redis client for the specified database number.
	//
	client := Database.DB(info.DbNum)

	// Parse out the key name
	//
	key := info.Key
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
func HandleCreateOperation(req *http.Request, info *RequestInfo) (response R) {
	e := "Create operations have not yet been implemented."
	response = R{"result": nil, "error": e}
	return
}

// Handles HTTP PUT requests, inteded for updating keys.
//
func HandleUpdateOperation(req *http.Request, info *RequestInfo) (response R) {
	client := Database.DB(info.DbNum)
	existsp, err := client.Exists(info.Key)
	if  err != nil {
		response = R{"result": nil, "error": err}
		return
	}
	if existsp {
		var errors []string

		// Check if the user specfieid an expiry time for the key.
		//
		if ttl := req.FormValue("ttl"); len(ttl) > 0 {
			ittl, err := strconv.Atoi(ttl)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s", err))
			}
			
			setp, err := client.Expire(info.Key, int64(ittl))
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s", err))
			}

			if setp {
				fmt.Println("EXPIRE", info.Key, ittl)
			}
		}

		// Get the value the user would like to set.
		//
		val := req.FormValue("value")
		fmt.Println("DEBUG", "Value =", val)

		// Now we need to branch, depending on the type of key we are setting
		// to.
		//
		keytype, err := client.Type(info.Key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s", err))
		}

		switch keytype {
		case "string":
			if offset := req.FormValue("offset"); len(offset) > 0 {
				i, err := strconv.Atoi(offset)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s", err))
				} else {
					err = client.Setrange(info.Key, offset, val)
					fmt.Println("SETRANGE", info.Key, offset, val)
				}
			}
			err = client.Set(info.Key, val)
			fmt.Println("SET", info.Key, val)

		case "set":
			err = client.Sadd(info.Key, val)
			fmt.Println("SADD", info.Key, val)

		case "zset":
			err = client.Zadd(info.Key, val)
			fmt.Println("ZADD", info.Key, val)

		case "hash":
			field := req.FormValue("field")
			if len(field) == 0 {
				e := "Missing required parameter: field."
				response = R{"result": nil, "error": e}
				return
			}
			err = client.Hset(info.Key, field, val)
			fmt.Println("HSET", info.Key, field, val)

		case "list":
			side := "right"
			if req.FormValue("side") == "left" {
				side := "left"
			}

			if side == "left" {
				err = client.Lpush(info.Key, val)
				fmt.Println("LPUSH", info.Key, val)
			} else {
				err = client.Rpush(info.Key, val)
				fmt.Println("RPUSH", info.Key, val)
			}
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("%s", err))
		}

		if len(errors) > 0 {
			response = R{"result": nil, "error": strings.Join(errors, " ")}
		} else {
			response = R{"result": true, "error": nil}
		}
	} else {
		e := "Key does not exist, cannot update."
		response = R{"result": nil, "error": e}
		return
	}
	return
}

func Favicon(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}
