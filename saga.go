package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Method 		string
	Url 		string
	Body 		[]byte
	Status 		int
}

type Saga struct {
	Leader 		string
	PartialReqs map[string]Request
	Status 		int
}

func getSagaFromReq(req *http.Request, leader string) Saga {
	defer req.Body.Close()

	body,_ := ioutil.ReadAll(req.Body)

	// TODO: construct Saga
	return Saga{
		Leader:      leader,
		PartialReqs: body,
	}
}

func (s *Saga)toByteArray() []byte {
	arr, _ := json.Marshal(s)
	return arr
}

func fromByteArray(arr []byte) *Saga {
	var s Saga
	_ = json.Unmarshal(arr, &s)
	return &s
}