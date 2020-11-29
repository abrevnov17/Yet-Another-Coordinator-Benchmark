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

	if server == ip {
		c.Redirect(http.StatusTemporaryRedirect, server)
		return
	}

	reqId := xid.New().String()
	sagas[reqId] = getSagaFromReq(c.Request, ip)

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
	reqId := c.Param("request")
	sagas[reqId] = getSagaFromReq(c.Request, c.Request.RemoteAddr)
	c.Status(http.StatusOK)
}

func partialRequestResponse(c *gin.Context) {
	reqId := c.Param("request")
	partial := c.Param("partial")
	// TODO: save partial response locally
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
	updateDisk(reqId, partial, sagas[reqId].PartialReqs[partial])
}

func delSaga(c *gin.Context) {
	reqId := c.Param("request")
	delete(sagas, reqId)
	removeFromDisk(reqId)
	c.Status(http.StatusOK)
}