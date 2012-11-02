// Provides functions related to reading values from Redis.
//
package main

import (
	"fmt"
	"net/http"
)

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
		response = R{"result": v.String(), "error": nil}

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

	case "none":
		response = R{"result": nil, "error": "Key does not exist."}

	default:
		e := fmt.Sprintf("Unknown type for key %s: %s", key, keyType)
		response = R{"result": nil, "error": e}
	}
	return
}
