package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(val)
	if err != nil {
		log.Printf("gob encode failed %v", err)
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})
	if err != nil {
		log.Printf("failed to publish: %v", err)
		return err
	}
	return nil
}

func SubscribeGob[T any](
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
			buf := bytes.NewBuffer(message.Body)
			decoder := gob.NewDecoder(buf)
			err := decoder.Decode(&m)
			if err != nil {
				log.Println(err)
				continue
			}

			ackType := handler(m)
			switch ackType {
			case Ack:
				message.Ack(false)
			case NackRequeue:
				message.Nack(false, true)
			case NackDiscard:
				message.Nack(false, false)
			}
		}
	}()
	return nil
}
