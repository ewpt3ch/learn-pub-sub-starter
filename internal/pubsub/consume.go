package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {

	qChannel, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	log.Printf("created queue %v\n", q)

	err = qChannel.Qos(10, 0, false)
	if err != nil {
		return err
	}

	consumeChan, err := qChannel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for message := range consumeChan {
			var m T
			err = json.Unmarshal(message.Body, &m)
			if err != nil {
				log.Println(err)
				continue
			}
			ackType := handler(m)
			switch ackType {
			case Ack:
				message.Ack(false)
				log.Println("Ack")
			case NackRequeue:
				message.Nack(false, true)
				log.Println("NackRequeue")
			case NackDiscard:
				message.Nack(false, false)
				log.Println("NackDiscard")
			}
		}
	}()
	return nil
}
