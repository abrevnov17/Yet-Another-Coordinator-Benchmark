package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Status int

const (
    Initialized Status = iota
    Running
    Failed
    Aborted
)

type Request struct {
	Method 		string `json:"method"`
	Url 		string `json:"url"`
	Body 		string `json:"body"`
	Status 		Status
}

type Saga struct {
	Leader 		string
	PartialReqs map[string]Request
	CompReqs map[string]Request
	Status 		Status // 1 if being processed, 2 on success, 3 on failure
}

type Transaction struct {
	PartialReqs map[string]Request `json:"partial_reqs"`
	CompReqs map[string]Request `json:"comp_reqs"`
}

/*
This creates a saga from a request. Requests should be of the following format (containing
two sets of key-value sets, one for partial requests and another for compensating requests):

{"partial_reqs": 
	{"some_partial_request_name": 
		{"method": "GET", "url": "/test/", "body": "test2"}
	}, 
"comp_reqs": 
	{"some_compensating_request_name": 
		{"method": "POST", "url": "/test2/", "body": "test3"}
	}
}
*/
func getSagaFromReq(req *http.Request, leader string) Saga {
	defer req.Body.Close()

	body,_ := ioutil.ReadAll(req.Body)

	var transaction Transaction

    // Decoded request body into transaction struct. If there is an error,
    // respond to the client with the error message and a 400 status code.
    err := json.NewDecoder(req.Body).Decode(&transaction)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	// construct Saga
	return Saga{
		Leader:      leader,
		PartialReqs: transaction.PartialReqs,
		CompReqs: transaction.CompReqs,
		Status: 1
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