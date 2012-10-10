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
	http.HandleFunc("/", dispatcher)
	http.HandleFunc("/favicon.ico", favicon)

	// Start listening for requests
	//
	println("Scarlet HTTP listening on", listenAddr)
	panic(http.ListenAndServe(listenAddr, nil))
	return
}

func dispatcher(rw http.ResponseWriter, req *http.Request) {
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
	println("DB #:", dbnum)

	// Parse out the key name
	//
	key := matches[3]
	println("Key:", key)

	var response R

	if len(key) == 0 {
		if req.Method == "GET" {
			// The length of the key name is zero, so just list all
			// of the keys in the database.
			//
			response = listKeys(dbnum)
		} else {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

	// Write out a response
	//
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, response)
	return
}

func favicon(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}

func listKeys(db int) (resp R) {
	keys, err := redisClient.Keys("*")
	if err != nil {
		resp = R{"result": nil, "error": err}
		return
	}
	resp = R{"result": keys, "error": nil}
	return
}
