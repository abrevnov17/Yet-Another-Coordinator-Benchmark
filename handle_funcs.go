package main

import (
	"github.com/gin-gonic/gin"
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

	// if server is not self
	c.Redirect(http.StatusTemporaryRedirect, server)
	// else
	// TODO: save saga locally

	// get sub cluster
	servers, ok := ring.GetNodes(key, subClusterSize)
	if ok == false {
		log.Fatal("Insufficient Correct Nodes")
	}

	// send request to sub cluster
	requestBody, _ := ioutil.ReadAll(c.Request.Body)
	ack := make(chan bool)
	for _,server := range servers {
		// TODO: add request identifier
		go sendPostMsg(server + "/saga/cluster", requestBody, ack)
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
		go sendPutMsg(server + "/saga/commit")
	}

	// TODO: execute partial requests

	// if success, reply success
	for _,server := range servers {
		go sendDelMsg(server + "/saga/cluster")
	}
	c.JSON(http.StatusOK, gin.H{})

	// else
	// TODO: send compensation requests to all partial requests

	c.JSON(http.StatusBadRequest, gin.H{})

}

func newSaga(c *gin.Context) {
	// TODO: save saga locally

	c.Status(http.StatusOK)
}

func partialRequestResponse(c *gin.Context) {
	// TODO: save partial response locally

	c.Status(http.StatusOK)
}

func commit(c *gin.Context) {
	partialRequest := c.Param("request")

	// TODO: flush partial request to disk
}

func delSaga(c *gin.Context) {
	// TODO: delete saga locally

	c.Status(http.StatusOK)
}