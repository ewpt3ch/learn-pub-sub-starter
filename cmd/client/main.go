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

	rabbitChannel, err := rClient.Channel()
	if err != nil {
		log.Fatalf(" rabbit channel failed: %v", err)
	}

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("issue getting username: %v", err)
	}

	gs := gamelogic.NewGameState(userName)

	// pause queue
	err = pubsub.SubscribeJSON(
		rClient,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("failed to create and subscribe to queue: %v", err)
	}

	// move queue
	err = pubsub.SubscribeJSON(
		rClient,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		func(move gamelogic.ArmyMove) {
			gs.HandleMove(move)
			fmt.Println("> ")
		},
	)
	if err != nil {
		log.Fatalf("failed to create and subscribe to move queue: %v", err)
	}

	// REPL loop
	for {

		commands := gamelogic.GetInput()
		if len(commands) == 0 {
			continue
		}
		command := commands[0]
		switch command {
		case "spawn":
			err := gs.CommandSpawn(commands)
			if err != nil {
				fmt.Printf("spawn error: %v", err)
				fmt.Println()
			}

		case "move":
			move, err := gs.CommandMove(commands)
			if err != nil {
				fmt.Printf("move failed: %v", err)
				break
			}

			err = pubsub.PublishJSON(
				rabbitChannel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+userName,
				move,
			)
			if err != nil {
				log.Printf("failed to publish  move: %v", err)
				continue
			}

			fmt.Println("move published")

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
