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
var conn, _, _ = zk.Connect([]string{"zookeeper"}, time.Second)

func main() {
	fmt.Println("starting up coordinator...")

	ip = os.Getenv("POD_IP") + ":8080"
	log.Println("pod ip: " + ip)

	conn, _, _ = zk.Connect([]string{"zookeeper"}, time.Second)

	//defer conn.Close()

	router := gin.Default()

	router.GET("/", welcome)
	router.POST("/saga/:request", processSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

	go checkIfNewLeader()
}