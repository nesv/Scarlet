// Provides functions that perform delete operations on Redis keys.
//
package main

import (
	"fmt"
	"net/http"
)

func HandleDeleteOperation(req *http.Request, info *RequestInfo) (response R) {
	client := Redis.Db(info.DbNum)
	v, err := client.Do("EXISTS", info.Key)
	if err != nil {
		response = R{"result": nil, "error": fmt.Sprintf("%s", err)}
		return
	}

	existsp, ok := v.(bool)
	if ok && existsp {
		// The key exists!
		//
		// NOW DELETE IT!
		//
		if _, err = client.Do("DEL", info.Key); err != nil {
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
