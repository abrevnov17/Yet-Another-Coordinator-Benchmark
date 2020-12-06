package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"log"
	"os"
	"time"
)

var ip string
var conn *zk.Conn

func main() {
	fmt.Println("starting up coordinator...")

	ip = os.Getenv("POD_IP") + ":8080"
	log.Println("pod ip: " + ip)

	conn, _, err := zk.Connect([]string{"zookeeper"}, time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	router := gin.Default()

	router.GET("/", welcome)
	router.POST("/saga", processSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

	go checkIfNewLeader()
}