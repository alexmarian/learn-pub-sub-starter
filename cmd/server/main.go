package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"log"
	"os"
	"os/signal"
	"syscall"
)
import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const amqpURI = "amqp://guest:guest@localhost:5672/"
	dial, err := amqp.Dial(amqpURI)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to RabbitMQ: %s", err))
	}
	defer dial.Close()
	fmt.Println("Connected to RabbitMQ")
	gamelogic.PrintServerHelp()
	channel, queue, err := pubsub.DeclareAndBind(dial, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.Durable)
	if err != nil {
		panic(fmt.Sprintf("Failed to declare and bind queue: %s", err))
	}
	defer channel.Close()
	fmt.Printf("Queue %s bound to exchange %s with key %s\n", queue.Name, routing.ExchangePerilTopic, "game_logs.*")
	for {
		input := gamelogic.GetInput()
		switch input[0] {

		case "pause":
			{
				log.Println("Pausing game...")
				err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
					IsPaused: true,
				})
				if err != nil {
					panic(fmt.Sprintf("Failed to publish message: %s", err))
				}
			}
		case "resume":
			{
				log.Println("Resuming game...")
				err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
					IsPaused: false,
				})
				if err != nil {
					panic(fmt.Sprintf("Failed to publish message: %s", err))
				}
			}
		case "quit":
			{
				break
			}
		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		<-signalChan
		fmt.Println("Shutting down gracefully...")
	}
}
