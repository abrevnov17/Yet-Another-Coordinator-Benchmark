package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
)

var zkEndpoints = []string {
	"zk-0.zk-hs.default.svc.cluster.local",
	"zk-1.zk-hs.default.svc.cluster.local",
	"zk-2.zk-hs.default.svc.cluster.local",
	"zk-3.zk-hs.default.svc.cluster.local",
	"zk-4.zk-hs.default.svc.cluster.local",
}

var ip string
var conn *zk.Conn

func main() {
	var err error

	fmt.Println("starting up coordinator...")

	ip = os.Getenv("POD_IP") + ":8080"
	log.Println("pod ip: " + ip)
	
	conn, _, err = zk.Connect(zkEndpoints, time.Second)
	if err != nil {
		log.Fatal(err)
	}

	//defer conn.Close()

	router := gin.Default()

	router.GET("/", welcome)
	router.POST("/saga/:request", processSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

	go checkIfNewLeader()
}
