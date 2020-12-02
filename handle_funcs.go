package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

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

	reqID := xid.New().String()

	saga, err := getSagaFromReq(c.Request, ip)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	sagas[reqID] = saga

	// get sub cluster
	servers, ok := ring.GetNodes(key, subClusterSize)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	// send request to sub cluster
	requestBody, _ := ioutil.ReadAll(c.Request.Body)
	ack := make(chan bool)
	for _, server := range servers {
		go sendPostMsg(server+"/saga/cluster/"+reqID, requestBody, ack)
	}

	// wait for majority of ack
	cnt := 0
	for cnt < len(coordinators)/2+1 {
		if <-ack {
			cnt++
		}
	}

	// broadcast commit
	for _, server := range servers {
		go sendPutMsg(server + "/saga/commit/" + reqID)
	}

	// execute partial requests
	rollbackTier, rollback := sendPartialRequests(saga)

	if rollback == true {
		// experiences failure, need to send compensating requests up to tier (inclusive)
		sendCompensatingRequests(saga, rollbackTier)
		c.JSON(http.StatusBadRequest, gin.H{})
	} else {
		// reply success
		for _, server := range servers {
			go sendDelMsg(server + "/saga/cluster/" + reqID)
		}
		c.JSON(http.StatusOK, gin.H{})
	}
}

func newSaga(c *gin.Context) {
	reqID := c.Param("request")
	sag, err := getSagaFromReq(c.Request, c.Request.RemoteAddr)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	sagas[reqID] = sag

	c.Status(http.StatusOK)
}

func partialRequestResponse(c *gin.Context) {
	reqID := c.Param("request")
	partial := c.Param("partial")
	// TODO: save partial response locally
	sagas[reqID].PartialReqs[partial] = Request{
		Method: "",
		Url:    "",
		Body:   nil,
	}
	c.Status(http.StatusOK)
}

func delSaga(c *gin.Context) {
	reqID := c.Param("request")
	delete(sagas, reqID)
	c.Status(http.StatusOK)
}
