package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"github.com/rs/xid"
	"log"
	"net/http"
)

func welcome(c *gin.Context) {
	fmt.Println("received a welcome message...")
	c.Status(http.StatusOK)
}

func processSaga(c *gin.Context) {
	saga, err := getSagaFromReq(c.Request, ip)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	reqID := xid.New().String()
	if _, err := conn.Create("/" + reqID, saga.toByteArray(), 0, zk.WorldACL(zk.PermAll)); err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// execute partial requests
	log.Println("Sending partial requests")
	rollbackTier, rollback := sendPartialRequests(reqID, &saga)

	if rollback == true {
		sendCompensatingRequests(reqID, rollbackTier, &saga)
		_ = conn.Delete("/" + reqID, 0)
		c.Status(http.StatusBadRequest)
	} else {
		_ = conn.Delete("/" + reqID, 0)
		c.Status(http.StatusOK)
	}
}
