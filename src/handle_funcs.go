package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func welcome(c *gin.Context) {
	fmt.Println("recieved a welcome message...")
	c.Status(http.StatusOK)
}

func processSaga(c *gin.Context) {
	key := c.Request.RemoteAddr
	server, ok := ring.GetNode(key)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	if server != ip {
		c.Redirect(http.StatusTemporaryRedirect, server)
		return
	}

	saga, err := getSagaFromReq(c.Request, ip)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	reqID := xid.New().String()
	sagasMutex.Lock()
	sagas[reqID] = saga
	sagasMutex.Unlock()

	// get sub cluster
	servers, ok := ring.GetNodes(key, subClusterSize)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	// send request to sub cluster
	ack := make(chan MsgStatus)
	for _, server := range servers {
		go sendPostMsg(server+"/saga/cluster/"+reqID, "", saga.toByteArray(), ack)
	}

	// wait for majority of ack
	cnt := 0
	for cnt < len(coordinators)/2+1 {
		if (<-ack).ok {
			cnt++
		}
	}

	// execute partial requests
	rollbackTier, rollback := sendPartialRequests(reqID, servers)

	if rollback == true {
		// experiences failure, need to send compensating requests up to tier (inclusive)
		sendCompensatingRequests(reqID, rollbackTier, servers)
		c.Status(http.StatusBadRequest)
	} else {
		// reply success
		for _, server := range servers {
			go sendDelMsg(server+"/saga/cluster/"+reqID, "", ack)
		}

		// wait for majority of ack
		cnt := 0
		for cnt < len(coordinators)/2+1 {
			if (<-ack).ok {
				cnt++
			}
		}
		// TODO: return with body
		c.Status(http.StatusOK)
	}
}

func newSaga(c *gin.Context) {
	reqID := c.Param("request")

	defer c.Request.Body.Close()
	body, _ := ioutil.ReadAll(c.Request.Body)
	saga := fromByteArray(body)

	sagasMutex.Lock()
	sagas[reqID] = saga
	sagasMutex.Unlock()

	c.Status(http.StatusOK)
}

func partialRequestResponse(c *gin.Context) {
	var resp PartialResponse
	var targetPartialRequest TransactionReq

	body, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	saga := sagas[resp.SagaId]
	saga.Leader = c.Request.RemoteAddr

	targetPartialRequest = saga.Transaction.Tiers[resp.Tier][resp.reqID]
	if resp.IsComp {
		targetPartialRequest.CompReq.Status = resp.Status
	} else {
		targetPartialRequest.PartialReq.Status = resp.Status
	}
	saga.Transaction.Tiers[resp.Tier][resp.ReqID] = targetPartialRequest

	sagasMutex.Lock()
	sagas[resp.SagaId] = saga
	sagasMutex.Unlock()

	c.Status(http.StatusOK)
}

func delSaga(c *gin.Context) {
	reqID := c.Param("request")
	delete(sagas, reqID)
	c.Status(http.StatusOK)
}
