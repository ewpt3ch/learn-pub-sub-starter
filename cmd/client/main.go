package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	conString := "amqp://guest:guest@localhost:5672/"
	rClient, err := amqp.Dial(conString)
	if err != nil {
		log.Fatalf("rabbitmq failure: %v", err)
	}
	defer rClient.Close()
	fmt.Println("rabbitMQ connection successful")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("issue getting username: %v", err)
	}

	qKey := "pause"
	qName := fmt.Sprintf("%s.%s", qKey, userName)

	_, _, err = pubsub.DeclareAndBind(rClient, "peril_direct", qName, qKey, pubsub.SimpleQueueTransient)
	if err != nil {
		fmt.Printf("declare and bind failed: %v", err)
	}

	fmt.Println("Starting Peril client...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	fmt.Printf("stopping client, %v", sig)

}
