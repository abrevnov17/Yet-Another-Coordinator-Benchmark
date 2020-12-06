package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
)

type MsgStatus struct {
	ok    bool
	reqID string
}

func getIpFromAddr(remoteAddr string) string {
	addrs := strings.Split(remoteAddr, ":")
	return addrs[0]
}

func sendPostMsg(url, reqID string, body []byte, ack chan MsgStatus) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqID: reqID}
}

func sendGetMsg(url, reqID string, ack chan MsgStatus) {
	resp, err := http.Get(url)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	log.Println("GET", url, err)
	ack <- MsgStatus{ok: ok, reqID: reqID}
}

func sendPutMsg(url, reqID string, body []byte, ack chan MsgStatus) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		ack <- MsgStatus{ok: false, reqID: reqID}
		return
	}

	resp, err := client.Do(req)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqID: reqID}
}

func sendDelMsg(url, reqID string, ack chan MsgStatus) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		ack <- MsgStatus{ok: false, reqID: reqID}
		return
	}
	resp, err := client.Do(req)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqID: reqID}
}

func sendMessage(reqID string, req Request, resp chan MsgStatus) {
	if req.Method == "POST" {
		body, err := json.Marshal(req.Body)
		if err != nil {
			resp <- MsgStatus{ok: false, reqID: reqID}
		}
		sendPostMsg(req.URL, reqID, body, resp)
	} else if req.Method == "GET" {
		sendGetMsg(req.URL, reqID, resp)
	} else if req.Method == "PUT" {
		body, err := json.Marshal(req.Body)
		if err != nil {
			resp <- MsgStatus{ok: false, reqID: reqID}
		}
		sendPutMsg(req.URL, reqID, body, resp)
	} else if req.Method == "DELETE" {
		sendDelMsg(req.URL, reqID, resp)
	} else {
		resp <- MsgStatus{ok: false, reqID: reqID}
	}
}

// Returns tier that needs to be rolled back to, nil on success
func sendPartialRequests(sagaId string, saga *Saga) (int, bool) {
	tiersMap := saga.Transaction.Tiers

	// since maps do not guarantee order, we get keys and sort them
	tiers := make([]int, 0, len(tiersMap))
	for tier := range tiersMap {
		tiers = append(tiers, tier)
	}

	sort.Ints(tiers)

	// loop in order of tier
	for _, tier := range tiers {
		requestsMap := tiersMap[tier]

		// iterate over requests and asynchronously send partial requests
		success := make(chan MsgStatus)
		for reqID, transaction := range requestsMap {
			go sendMessage(reqID, transaction.PartialReq, success)
		}

		// wait for successes from each partial request, quit on failure
		cnt := 0
		for cnt < len(requestsMap) {
			status := <-success
			if status.ok {
				request := tiersMap[tier][status.reqID]
				request.PartialReq.Status = Success
				saga.Transaction.Tiers[tier][status.reqID] = request
				_, _ = conn.Set("/" + sagaId, saga.toByteArray(), 0)
				cnt++
			} else {
				// failure at this tier, need to roll back
				request := tiersMap[tier][status.reqID]
				request.PartialReq.Status = Failed
				saga.Transaction.Tiers[tier][status.reqID] = request
				_, _ = conn.Set("/" + sagaId, saga.toByteArray(), 0)
				return tier, true
			}
		}
	}

	return -1, false
}

// Tier is highest tier through which (inclusive) we need to roll back
func sendCompensatingRequests(sagaId string, maxTier int, saga *Saga) {
	tiersMap := saga.Transaction.Tiers

	// since maps do not guarantee order, we get keys and sort them
	tiers := make([]int, 0, len(tiersMap))
	for tier := range tiersMap {
		tiers = append(tiers, tier)
	}

	sort.Ints(tiers)

	// loop in order of tier
	for _, tier := range tiers {
		// we have rolled back all the requests necessary

		if tier > maxTier {
			return
		}
		requestsMap := tiersMap[tier]

		// iterate over requests and asynchronously send partial requests
		success := make(chan MsgStatus)
		for reqID, transaction := range requestsMap {
			go sendMessage(reqID, transaction.CompReq, success)
		}

		// wait for successes from each partial request, quit on failure
		cnt := 0
		for cnt < len(requestsMap) {
			status := <-success
			if status.ok {
				request := tiersMap[tier][status.reqID]
				request.PartialReq.Status = Aborted
				request.CompReq.Status = Success
				saga.Transaction.Tiers[tier][status.reqID] = request
				_, _ = conn.Set("/" + sagaId, saga.toByteArray(), 0)
				cnt++
			}
		}
	}
}


func checkIfNewLeader() {
	// TODO: get all requests
}
