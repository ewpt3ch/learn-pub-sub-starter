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

	_, q, err := pubsub.DeclareAndBind(rClient, routing.ExchangePerilDirect, routing.GameLogSlug, "game_logs.*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("error creating queue: %v", err)
	}
	fmt.Printf("queue %v created", q)

	gamelogic.PrintServerHelp()

	for {
		commands := gamelogic.GetInput()
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
