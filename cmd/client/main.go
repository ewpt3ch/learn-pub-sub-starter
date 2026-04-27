package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
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

	qKey := "pause"
	qName := fmt.Sprintf("%s.%s", qKey, userName)

	_, queue, err := pubsub.DeclareAndBind(rClient, "peril_direct", qName, qKey, pubsub.SimpleQueueTransient)
	if err != nil {
		fmt.Printf("declare and bind failed: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gameState := gamelogic.NewGameState(userName)

	for {

		commands := gamelogic.GetInput()
		command := commands[0]
		switch command {
		case "spawn":
			err := gameState.CommandSpawn(commands)
			if err != nil {
				fmt.Printf("spawn error: %v", err)
				break
			}

		case "move":
			army, err := gameState.CommandMove(commands)
			if err != nil {
				fmt.Printf("move failed: %v", err)
				break
			}

			fmt.Printf("Army %v moved\n", army)

		case "status":
			gameState.CommandStatus()

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
