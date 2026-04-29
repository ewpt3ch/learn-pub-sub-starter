package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	// logs queue
	err = pubsub.SubscribeGob(
		rClient,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("error creating queue: %v", err)
	}
	fmt.Println("log queue created")

	gamelogic.PrintServerHelp()

	for {
		commands := gamelogic.GetInput()
		if len(commands) == 0 {
			continue
		}
		command := commands[0]
		switch command {
		case "pause":
			fmt.Println("sending pause message")
			err = pubsub.PublishJSON(rabbitChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				fmt.Printf("failed to send message: %v", err)
			}

		case "resume":
			fmt.Println("resuming game")
			err = pubsub.PublishJSON(rabbitChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})

		case "quit":
			fmt.Println("quiting")
			return

		default:
			fmt.Println("command not understood")
			continue

		}

	}

}
