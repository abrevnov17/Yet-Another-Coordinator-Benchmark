package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"log"
	"net/http"
)

func welcome(c *gin.Context) {
	fmt.Println("received a welcome message...")
	c.Status(http.StatusOK)
}

func processSaga(c *gin.Context) {
	sagaId := c.Param("request")
	saga, err := getSagaFromReq(c.Request, ip)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if _, err := conn.Create("/" + sagaId, saga.toByteArray(), int32(0), zk.WorldACL(zk.PermAll)); err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// execute partial requests
	log.Println("Sending partial requests")
	rollbackTier, rollback := sendPartialRequests(sagaId, &saga)

	if rollback == true {
		log.Println("Sending compensation requests")
		sendCompensatingRequests(sagaId, rollbackTier, &saga)
		log.Println("Deleting saga")
		_ = conn.Delete("/" + sagaId, 0)
		c.Status(http.StatusBadRequest)
	} else {
		log.Println("Deleting saga")
		_ = conn.Delete("/" + sagaId, 0)
		c.Status(http.StatusOK)
	}
}
