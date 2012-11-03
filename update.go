// Provides functions related to updating existing keys, in Redis.
//
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Handles HTTP PUT requests, inteded for updating keys.
//
func HandleUpdateOperation(req *http.Request, info *RequestInfo) (response R) {
	client, err := Database.DB(info.DbNum)
	v, err := client.Do("EXISTS", info.Key)
	if err != nil {
		response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		return
	}
	existsp, ok := v.(bool)
	if ok && existsp {
		var errors []string

		// Check if the user specfieid an expiry time for the key.
		//
		if ttl := req.FormValue("ttl"); len(ttl) > 0 {
			ittl, err := strconv.Atoi(ttl)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s", err))
			}

			v, err := client.Do("EXPIRE", info.Key, int64(ittl))
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s", err))
			}

			setp, ok := v.(bool)
			if ok && setp {
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
		v, err := client.Do("TYPE", info.Key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s", err))
		}

		keytype, ok := v.(string)
		if !ok {
			msg := "Could not convert reply to string."
			response = R{"result": nil, "error": msg}
			return
		}
		switch keytype {
		case "string":
			if offset := req.FormValue("offset"); len(offset) > 0 {
				i, err := strconv.Atoi(offset)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s", err))
				} else {
					_, err = client.Do("SETRANGE", info.Key, i, val)
					fmt.Println("SETRANGE", info.Key, i, val)
				}
			}
			_, err = client.Do("SET", info.Key, val)
			fmt.Println("SET", info.Key, val)

		case "set":
			_, err = client.Do("SADD", info.Key, val)
			fmt.Println("SADD", info.Key, val)

		case "zset":
			var ranking float64 = 1.0
			if v := req.FormValue("ranking"); len(v) > 0 {
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s", err))
				} else {
					ranking = f
				}
			}
			_, err = client.Do("ZADD", info.Key, ranking, val)
			fmt.Println("ZADD", info.Key, ranking, val)

		case "hash":
			field := req.FormValue("field")
			if len(field) == 0 {
				e := "Missing required parameter: field."
				response = R{"result": nil, "error": e}
				return
			}
			_, err = client.Do("HSET", info.Key, field, val)
			fmt.Println("HSET", info.Key, field, val)

		case "list":
			side := "right"
			if req.FormValue("side") == "left" {
				side = "left"
			}

			if side == "left" {
				_, err = client.Do("LPUSH", info.Key, val)
				fmt.Println("LPUSH", info.Key, val)
			} else {
				_, err = client.Do("RPUSH", info.Key, val)
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
