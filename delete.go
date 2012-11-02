// Provides functions that perform delete operations on Redis keys.
//
package main

import (
	"fmt"
	"net/http"
)

func HandleDeleteOperation(req *http.Request, info *RequestInfo) (response R) {
	client := Database.DB(info.DbNum)
	existsp, err := client.Exists(info.Key)
	if err != nil {
		response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		return
	}

	if existsp {
		// The key exists!
		//
		// NOW DELETE IT!
		//
		if _, err = client.Del(info.Key); err != nil {
			response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		} else {
			fmt.Println("DEL", info.Key)
			response = R{"result": true, "error": nil}
		}
	} else {
		response = R{"result": nil, "error": "Key does not exist."}
	}
	return
}
