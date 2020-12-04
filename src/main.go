package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/serialx/hashring"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var coordinators = []string{
	"localhost: 8080",
	"192.168.0.2: 8000",
	"192.168.0.3: 1234",
}

var ring = hashring.New(coordinators)

var ip = "localhost:8080"

var clientset *kubernetes.Clientset

func main() {
	// creating k8s client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// TODO: update self IP

	// TODO: update coordinators

	ring = hashring.New(coordinators)

	go updateCoordinatorsList()

	router := gin.Default()

	router.POST("/saga", processSaga)
	router.POST("/saga/cluster/:request", newSaga)
	router.PUT("/saga/partial", partialRequestResponse)
	router.DELETE("/saga/:request", delSaga)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

func updateCoordinatorsList() {
	for {
		time.Sleep(500 * time.Millisecond)

		coordinators = pullCoordinators()
		// update ring
		ring = hashring.New(coordinators)
		checkIfNewLeader()
	}
}

func pullCoordinators() []string {
	// pull coordinators from k8s
	pods, err := clientset.CoreV1().Pods("yac").List(metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	// TESTING...REMOVE LATER
	for _, pod := range pods.Items {
		fmt.Println(json.Marshal(pod))
	}

	// TODO: return actual vlaue
	return nil
}
