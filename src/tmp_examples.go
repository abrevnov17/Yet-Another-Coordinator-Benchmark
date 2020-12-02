package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type Resp struct {
	Status		int
	Context 	string
}

type TmpMessage struct {
	Value 		string
}

var chanMap = make(map[int]chan Resp)

func sendPostRequest(url string, msg TmpMessage) {
	jsonStr,_ := json.Marshal(msg)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))

	if err != nil {
		panic(err)
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Println(resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &msg); err != nil {
		panic(err)
	}

	chanMap[0] <- Resp{Status:200, Context:msg.Value}

	if err = resp.Body.Close(); err != nil {
		panic(err)
	}
}

func example1(c *gin.Context) {
	var msg TmpMessage

	huat, _ := ioutil.ReadAll(c.Request.Body)
	if err := json.Unmarshal(huat, &msg); err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, msg)
}

func example2(c *gin.Context)  {
	ip := c.Request.RemoteAddr
	servers,_ :=  ring.GetNodes("key", 2)
	resp := make(chan Resp)

	chanMap[0] = resp
	go sendPostRequest("http://localhost:8080/pong", TmpMessage{Value:"value"})

	c.JSON(200, gin.H{
		"sender_addr": ip,
		"hashring":    servers,
		"param": c.Param("param"),
		"chan": <-resp,
	})
}
