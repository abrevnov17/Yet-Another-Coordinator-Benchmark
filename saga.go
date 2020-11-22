package main

import (
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

func getSagaFromReq(req *http.Request) Saga {
	defer req.Body.Close()

	body,_ := ioutil.ReadAll(req.Body)

	// TODO: construct Saga
	return Saga{
		Leader:      "",
		PartialReqs: body,
	}
}

func getSagaFromLeaderReq(req *http.Request) Saga {
	defer req.Body.Close()

	body,_ := ioutil.ReadAll(req.Body)

	// TODO: construct Saga
	return Saga{
		Leader:      "",
		PartialReqs: body,
	}
}