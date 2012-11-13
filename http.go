// Provides the HTTP interface for Scarlet.
//
package main

import (
	"errors"
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
	info, err := Redis.Info("all")
	response = R{"result": info, "error": err}
	fmt.Fprint(rw, response)
	return
}

type RequestInfo struct {
	DbNum int
	Key   string
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

func RootHandler(req *http.Request) (response R) {
	// Info on our host's immediate slaves.
	//
	var slaveInfo []string
	for _, slave := range Redis.Slaves {
		slaveInfo = append(slaveInfo, slave.Addr)
	}

	// Who is our master?
	//
	var masterAddress string
	if info, err := Redis.Info("replication"); err == nil {
		Debug(fmt.Sprint("Role: ", info["role"]))
		switch info["role"] {
		case "master":
			masterAddress = ""
		case "slave":
			masterAddress = fmt.Sprintf("%s:%s", info["master_host"], info["master_port"])
		}
	}

	// Finally, build up the response, and return.
	//
	response = R{"result": R{"databases": Redis.NConnections(),
		"slaves": slaveInfo, "master": masterAddress}, "error": nil}
	return
}

// Dispatches the incoming request to the proper action handler, depending on
// the HTTP method that was used.
//
func DispatchRequest(rw http.ResponseWriter, req *http.Request) {
	// Who doesn't like logging?
	//
	fmt.Printf("HTTP\t%s\t%s\t%s\n", req.RemoteAddr, req.Method, req.URL.String())

	// Now, dispatch the request to the appropriate handler function.
	//
	var response R
	if req.URL.String() == "/" {
		response = RootHandler(req)
	} else if info, err := GetRequestInfo(req); err == nil {
		switch req.Method {
		case "GET":
			response = HandleReadOperation(req, info)

		case "POST":
			response = HandleCreateOperation(req, info)

		case "PUT":
			response = HandleUpdateOperation(req, info)

		case "DELETE":
			response = HandleDeleteOperation(req, info)
		}
	} else {
		response = R{"result": nil, "error": err}
	}
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, response)
	return
}

func Favicon(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}
