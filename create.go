// Provides functions related to creating new Redis keys.
//
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Handles HTTP POST requests, intended for creating new keys.
//
func HandleCreateOperation(req *http.Request, info *RequestInfo) (response R) {
	client, err := Database.DB(info.DbNum)
	v, err := client.Do("EXISTS", info.Key)
	if err != nil {
		response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		return
	}

	// Does the key already exist? If so, bomb out. We only want to be able
	// to create keys from HTTP POST requests.
	//
	existsp, ok := v.(bool)
	if ok && existsp {
		response = R{"result": nil, "error": "Key already exists."}
		return
	}

	// Oooh, the key doesn't exist?! Delicious.
	//
	// Let's see if the user explicitly stated the type of key to create.
	// If they didn't, then we will default to just using a string. Mind you,
	// if they did something silly like specify an unsupported type, then
	// we should just bomb out, letting them know they made a boo-boo.
	//
	var keytype string
	if ktype := req.FormValue("type"); len(ktype) > 0 {
		switch ktype {
		case "list":
			fallthrough
		case "set":
			fallthrough
		case "zset":
			fallthrough
		case "hash":
			fallthrough
		case "string":
			keytype = ktype
		default:
			response = R{"result": nil, "error": "Invalid key type."}
		}
	} else {
		// Ahh, in the event the caller did not explicitly try and specify
		// a type for the new key, just default to "string".
		//
		keytype = "string"
	}

	// Let's just quickly make sure the user actually supplied a value to
	// be set.
	//
	value := req.FormValue("value")
	if len(value) == 0 {
		response = R{"result": nil, "error": "No value provided."}
		return
	}

	// Now, it's time to switch which command we use depending on the key
	// type.
	//
	switch keytype {
	case "string":
		_, err = client.Do("SET", info.Key, value)

	case "list":
		_, err = client.Do("LPUSH", info.Key, value)

	case "set":
		_, err = client.Do("SADD", info.Key, value)

	case "zset":
		var ranking float64
		if rv := req.FormValue("ranking"); len(rv) > 0 {
			ranking, err = strconv.ParseFloat(rv, 64)
			if err != nil {
				response = R{"result": nil, "error": err}
				return
			}
		} else {
			ranking = 1.0
		}
		_, err = client.Do("ZADD", info.Key, ranking, value)

	case "hash":
		if field := req.FormValue("field"); len(field) > 0 {
			_, err = client.Do("HSET", info.Key, field, value)
		} else {
			err = errors.New("No field name specified.")
		}
	}

	// If any errors cropped up, mark the call as a failure and provide an
	// error.
	//
	if err != nil {
		response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
	} else {
		response = R{"result": true, "error": nil}
	}
	return
}
