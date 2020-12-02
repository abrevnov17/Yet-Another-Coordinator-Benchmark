package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
)

type MsgStatus struct {
	ok 		bool
	reqId 	string
}

func sendPostMsg(url, reqId string, body []byte, ack chan MsgStatus) {
	resp, err := http.Post("http://"+url, "application/json", bytes.NewBuffer(body))
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqId: reqId}
}

func sendGetMsg(url, reqId string, ack chan MsgStatus) {
	resp, err := http.Get("http://"+url)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqId: reqId}
}

func sendPutMsg(url, reqId string, body []byte, ack chan MsgStatus) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "http://"+url, bytes.NewBuffer(body))
	if err != nil {
		ack <- MsgStatus{ok: false, reqId: reqId}
		return
	}

	resp, err := client.Do(req)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqId: reqId}
}

func sendDelMsg(url, reqId string, ack chan MsgStatus) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://"+url, nil)
	if err != nil {
		ack <- MsgStatus{ok: false, reqId: reqId}
		return
	}
	resp, err := client.Do(req)
	ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299
	ack <- MsgStatus{ok: ok, reqId: reqId}
}

func sendMessage(reqId string, req Request, resp chan MsgStatus) {
	if req.Method == "POST" {
		body, err := json.Marshal(req.Body)
		if err != nil {
			resp <- MsgStatus{ok: false, reqId: reqId}
		}
		sendPostMsg(req.URL, reqId, body, resp)
	} else if req.Method == "GET" {
		sendGetMsg(req.URL, reqId, resp)
	} else if req.Method == "PUT" {
		body, err := json.Marshal(req.Body)
		if err != nil {
			resp <- MsgStatus{ok: false, reqId: reqId}
		}
		sendPutMsg(req.URL, reqId, body, resp)
	} else if req.Method == "DELETE" {
		sendDelMsg(req.URL, reqId, resp)
	} else {
		resp <- MsgStatus{ok: false, reqId: reqId}
	}
}

// Returns tier that needs to be rolled back to, nil on success
func sendPartialRequests(sagaId string, subCluster []string) (int, bool) {
	tiersMap := sagas[sagaId].Transaction.Tiers

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
		for reqId, transaction := range requestsMap {
			go sendMessage(reqId, transaction.PartialReq, success)
		}

		// wait for successes from each partial request, quit on failure
		// TODO: add retries / timeouts
		cnt := 0
		for cnt < len(requestsMap) {
			status := <- success
			if status.ok {
				updateSubCluster(sagaId, tier, status.reqId, false, Success, subCluster)
				cnt++
			} else {
				// failure at this tier, need to roll back
				updateSubCluster(sagaId, tier, status.reqId, false, Failed, subCluster)
				return tier, true
			}
		}
	}

	return -1, false
}

// Tier is highest tier through which (inclusive) we need to roll back
func sendCompensatingRequests(sagaId string, maxTier int, subCluster []string) {
	tiersMap := sagas[sagaId].Transaction.Tiers

	// since maps do not guarantee order, we get keys and sort them
	tiers := make([]int, 0, len(tiersMap))
	for tier := range tiersMap {
		tiers = append(tiers, tier)
	}

	sort.Ints(tiers)

	// loop in order of tier
	for _, tier := range tiers {
		// we have rolled back all the requests necessary

		if tier >= maxTier {
			return
		}
		requestsMap := tiersMap[tier]

		// iterate over requests and asynchronously send partial requests
		success := make(chan MsgStatus)
		for reqId, transaction := range requestsMap {
			go sendMessage(reqId, transaction.CompReq, success)
		}

		// wait for successes from each partial request, quit on failure
		// TODO: add retries / timeouts
		cnt := 0
		for cnt < len(requestsMap) {
			status := <- success
			if status.ok {
				updateSubCluster(sagaId, tier, status.reqId, true, Success, subCluster)
				cnt++
			}
			// TODO: handle unsuccessful compensation
		}
	}
}

func updateSubCluster(sagaId string, tier int, reqId string, isComp bool, status Status, subCluster []string) {
	body,_ := json.Marshal(PartialResponse{
		SagaId: sagaId,
		Tier:   tier,
		ReqId:  reqId,
		IsComp: isComp,
		Status: status,
	})

	ack := make(chan MsgStatus)
	for _, server := range subCluster {
		go sendPutMsg(server + "/saga/partial", "", body, ack)
	}

	cnt := 0
	for cnt < len(subCluster)/2+1 {
		if (<-ack).ok {
			cnt++
		}
	}
}

func checkIfNewLeader() {
	coordinatorSet := make(map[string]bool, len(coordinators))
	for _, c := range coordinators {
		coordinatorSet[c] = true
	}

	for id, s := range sagas {
		if _, isIn := coordinatorSet[s.Leader]; !isIn {
			newLeader, _ := ring.GetNode(id)
			s.Leader = newLeader
			sagas[id] = s
			if newLeader == ip {
				// TODO: send compensation requests
			}
		}
	}
}
