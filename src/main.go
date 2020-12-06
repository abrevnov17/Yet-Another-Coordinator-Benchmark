package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
)

var ip string
var conn *zk.Conn

func main() {
	fmt.Println("starting up coordinator...")

	ip = os.Getenv("POD_IP") + ":8080"
	log.Println("pod ip: " + ip)

	conn, _, _ = zk.Connect([]string{"zk-0.zk-hs.default.svc.cluster.local",
		"zk-1.zk-hs.default.svc.cluster.local",
		"zk-2.zk-hs.default.svc.cluster.local"}, time.Second)

	router := gin.Default()

	router.GET("/", welcome)
	router.POST("/saga", processSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

	go checkIfNewLeader()
}
