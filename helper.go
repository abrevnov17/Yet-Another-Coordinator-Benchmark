package main

import (
	"bytes"
	"net/http"
)

func sendPostMsg(url string, body []byte, ack chan bool) {
	resp, err := http.Post("http://" + url, "application/json", bytes.NewBuffer(body))
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		ack <- true
	} else {
		ack <- false
	}
}

func sendPutMsg(url string) {
	client := &http.Client{}

	req, _ := http.NewRequest("PUT", "http://" + url, nil)

	_, _ = client.Do(req)
}

func sendDelMsg(url string) {
	client := &http.Client{}

	req, _ := http.NewRequest("DELETE", "http://" + url, nil)

	resp, err := client.Do(req)
	for err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp, err = client.Do(req)
	}
}

func sendPartialRequests(){
	// TODO: define partial requests format
}

func sendCompensation() {
	// TODO: send compensation requests
}

func checkIfNewLeader() {
	// TODO: iterate through sagas and if a saga's leader is down

	// TODO: if self is the new leader, send compensation requests
}