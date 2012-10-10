package main

import "encoding/json"

type R map[string]interface{}

func (r R) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}
