package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
)

func sendPostMsg(url string, body []byte, ack chan bool) {
	resp, err := http.Post("http://"+url, "application/json", bytes.NewBuffer(body))
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		ack <- true
	} else {
		ack <- false
	}
}

func sendMessage(req Request, resp chan bool) error {
	if req.Method == "POST" {
		bytes, err := json.Marshal(req.Body)
		if err != nil {
			return err
		}
		sendPostMsg(req.URL, bytes, resp)
	} else {
		return errors.New("Unsupported request method")
	}

	return nil
}

func sendPutMsg(url string) {
	client := &http.Client{}

	req, _ := http.NewRequest("PUT", "http://"+url, nil)

	_, _ = client.Do(req)
}

func sendDelMsg(url string) {
	client := &http.Client{}

	req, _ := http.NewRequest("DELETE", "http://"+url, nil)

	resp, err := client.Do(req)
	for err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp, err = client.Do(req)
	}
}

// Returns tier that needs to be rolled back to, nil on success
func sendPartialRequests(saga Saga) (int, bool) {
	tiersMap := saga.Transaction.Tiers

	// since maps do not guarentee order, we get keys and sort them
	tiers := make([]int, 0, len(tiersMap))
	for tier := range tiersMap {
		tiers = append(tiers, tier)
	}

	sort.Ints(tiers)

	// loop in order of tier
	for _, tier := range tiers {
		requestsMap := tiersMap[tier]

		// iterate over requests and asyncronously send partial requests
		success := make(chan bool)
		for _, transaction := range requestsMap {
			go sendMessage(transaction.PartialReq, success)
		}

		// wait for successes from each partial request, quit on failure
		// TODO: add retries / timeouts
		cnt := 0
		for cnt < len(requestsMap) {
			if <-success {
				cnt++
			} else {
				// failure at this tier, need to roll back
				return tier, true
			}
		}
	}

	return -1, false
}

// Tier is highest tier through which (inclusive) we need to roll back
func sendCompensatingRequests(saga Saga, maxTier int) {
	tiersMap := saga.Transaction.Tiers

	// since maps do not guarentee order, we get keys and sort them
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

		// iterate over requests and asyncronously send partial requests
		success := make(chan bool)
		for _, transaction := range requestsMap {
			go sendMessage(transaction.CompReq, success)
		}

		// wait for successes from each partial request, quit on failure
		// TODO: add retries / timeouts
		cnt := 0
		for cnt < len(requestsMap) {
			if <-success {
				cnt++
			}
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
