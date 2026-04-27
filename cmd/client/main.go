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

	fmt.Println("Starting Peril client...")

	conString := "amqp://guest:guest@localhost:5672/"
	rClient, err := amqp.Dial(conString)
	if err != nil {
		log.Fatalf("rabbitmq failure: %v", err)
	}
	defer rClient.Close()
	fmt.Println("rabbitMQ connection successful")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("issue getting username: %v", err)
	}

	gs := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(
		rClient,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs))
	if err != nil {
		log.Fatalf("failed to create and subscribe to queue: %v", err)
	}

	for {

		commands := gamelogic.GetInput()
		command := commands[0]
		switch command {
		case "spawn":
			err := gs.CommandSpawn(commands)
			if err != nil {
				fmt.Printf("spawn error: %v", err)
			}

		case "move":
			army, err := gs.CommandMove(commands)
			if err != nil {
				fmt.Printf("move failed: %v", err)
				break
			}

			fmt.Printf("Army %v moved\n", army)

		case "status":
			gs.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("command not understood")
		}

	}

}
