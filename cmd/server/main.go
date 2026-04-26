package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

	fmt.Println("Starting Peril server...")

	rabbitChannel, err := rClient.Channel()
	if err != nil {
		log.Fatalf("rabbit channel failed: %v", err)
	}

	err = pubsub.PublishJSON(rabbitChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})
	if err != nil {
		fmt.Printf("failed to send message: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	fmt.Printf("stopping server, %v", sig)

}
