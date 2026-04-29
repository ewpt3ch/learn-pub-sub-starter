package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
		handlerMove(gs, rabbitChannel),
	)
	if err != nil {
		log.Fatalf("failed to create and subscribe to move queue: %v", err)
	}

	//war queue
	err = pubsub.SubscribeJSON(
		rClient,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, rabbitChannel),
	)
	if err != nil {
		log.Fatalf("failed to create war queue %v", err)
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
			fmt.Println("Spamming destroys servers!")
			if len(commands) != 2 {
				fmt.Println("need 2 words")
				continue
			}
			spamN, err := strconv.Atoi(commands[1])
			if err != nil {
				fmt.Println("need a number: spam 1000")
				continue
			}
			for _ = range spamN {
				malLog := gamelogic.GetMaliciousLog()
				gl := routing.GameLog{
					CurrentTime: time.Now(),
					Message:     malLog,
					Username:    gs.GetUsername(),
				}
				pubsub.PublishGob(
					rabbitChannel,
					routing.ExchangePerilTopic,
					routing.GameLogSlug+"."+gs.GetUsername(),
					gl,
				)
			}

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("command not understood")
		}

	}

}
