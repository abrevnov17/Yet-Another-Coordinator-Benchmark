package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"net/http"
)

func processSaga(c *gin.Context) {
	key := c.Request.RemoteAddr
	server, ok := ring.GetNode(key)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	// TODO: if server is not self
	c.Redirect(http.StatusTemporaryRedirect, server)
	// TODO: else
	reqId := xid.New().String()
	sagas[reqId] = getSagaFromReq(c.Request)

	// get sub cluster
	servers, ok := ring.GetNodes(key, subClusterSize)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	// send request to sub cluster
	requestBody, _ := ioutil.ReadAll(c.Request.Body)
	ack := make(chan bool)
	for _,server := range servers {
		go sendPostMsg(server + "/saga/cluster/" + reqId, requestBody, ack)
	}

	// wait for majority of ack
	cnt := 0
	for cnt < len(coordinators) / 2 + 1 {
		if <-ack {
			cnt += 1
		}
	}

	// broadcast commit
	for _,server := range servers {
		// TODO: add request identifier
		go sendPutMsg(server + "/saga/commit/" + reqId)
	}

	// TODO: execute partial requests

	// if success, reply success
	for _,server := range servers {
		go sendDelMsg(server + "/saga/cluster/" + reqId)
	}
	c.JSON(http.StatusOK, gin.H{})

	// else
	// TODO: send compensation requests to all partial requests

	c.JSON(http.StatusBadRequest, gin.H{})

}

func newSaga(c *gin.Context) {
	// TODO: save saga locally
	reqId := c.Param("request")
	sagas[reqId] = getSagaFromLeaderReq(c.Request)
	c.Status(http.StatusOK)
}

func partialRequestResponse(c *gin.Context) {
	// TODO: save partial response locally
	reqId := c.Param("request")
	partial := c.Param("partial")
	sagas[reqId].PartialReqs[partial] = Request{
		Method: "",
		Url:    "",
		Body:   nil,
	}
	c.Status(http.StatusOK)
}

func commit(c *gin.Context) {
	reqId := c.Param("request")
	partial := c.Param("partial")

	// TODO: flush partial request to disk
	_ = sagas[reqId].PartialReqs[partial]
}

func delSaga(c *gin.Context) {
	// TODO: delete saga locally
	delete(sagas, c.Param("request"))
	c.Status(http.StatusOK)
}