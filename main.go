package main

import (
	"github.com/gin-gonic/gin"
	"github.com/serialx/hashring"
	"time"
)

var coordinators = []string{
	"localhost: 8080",
	"192.168.0.2: 8000",
	"192.168.0.3: 1234",
}

var sagas = make(map[string]Saga)

var ring = hashring.New(coordinators)

func main() {
	go updateCoordinatorsList()

	router := gin.Default()

	router.POST("/saga", processSaga)
	router.POST("/saga/cluster/:request", newSaga)
	router.PUT("/saga/partial", partialRequestResponse)
	router.PUT("/saga/commit/:request/:partial", commit)
	router.DELETE("/saga/cluster/:request", delSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

func updateCoordinatorsList() {
	for {
		time.Sleep(500 * time.Millisecond)
		// TODO: pull latest coordinators from k8s

		// update ring
		ring = hashring.New(coordinators)
		checkIfNewLeader()
	}
}