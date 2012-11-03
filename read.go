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
	client, err := Database.DB(info.DbNum)

	// Parse out the key name
	//
	key := info.Key
	if len(key) == 0 {
		// The length of the key name is zero, so just list all
		// of the keys in the database.
		//
		fmt.Println("KEYS", "*")
		v, err := client.Do("KEYS", "*")
		if err != nil {
			response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		} else {
			keys, ok := v.([]string)
			if !ok {
				msg := "Could not convert to string array."
				response = R{"result": nil, "error": msg}
				return
			}
			response = R{"result": keys, "error": nil}
		}
		return
	}

	// Get the key type, so that we know how to properly format the
	// response.
	//
	v, err := client.Do("TYPE", key)
	if err != nil {
		response = R{"result": nil, "error": err}
		return
	}

	keyType, ok := v.(string)
	if !ok {
		msg := fmt.Sprintf("Could not convert %v to string.", v)
		response = R{"result": nil, "error": msg}
		return
	}

	// Format the response according to the type the key holds.
	//
	switch keyType {
	case "string":
		println("GET", key)
		v, _ := client.Do("GET", key)
		r, _ := v.(string)
		response = R{"result": r, "error": nil}

	case "set":
		println("SMEMBERS", key)
		v, _ := client.Do("SMEMBERS", key)
		r, _ := v.([]string)
		response = R{"result": r, "error": nil}

	case "zset":
		println("ZRANGE", key, 0, -1)
		v, _ := client.Do("ZRANGE", key, 0, -1)
		r, _ := v.([]string)
		response = R{"result": r, "error": nil}

	case "list":
		println("LRANGE", key, 0, -1)
		v, _ := client.Do("LRANGE", key, 0, -1)
		r, _ := v.([]string)
		response = R{"result": r, "error": nil}

	case "hash":
		if field := req.FormValue("field"); field != "" {
			println("HGET", key, field)
			v, _ := client.Do("HGET", key, field)
			r, _ := v.(string)
			response = R{"result": r, "error": nil}
		} else {
			println("HGETALL", key)
			v, _ := client.Do("HGETALL", key)
			r, _ := v.(map[string]string)
			response = R{"result": r, "error": nil}
		}

	case "none":
		response = R{"result": nil, "error": "Key does not exist."}

	default:
		e := fmt.Sprintf("Unknown type for key %s: %s", key, keyType)
		response = R{"result": nil, "error": e}
	}
	return
}
