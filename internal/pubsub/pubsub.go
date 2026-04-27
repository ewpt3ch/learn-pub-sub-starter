package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	})
	if err != nil {
		return err
	}

	return nil

}

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {

	queueChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	isDurable := queueType == SimpleQueueDurable
	isTransient := queueType == SimpleQueueTransient

	queue, err := queueChannel.QueueDeclare(queueName, isDurable, isTransient, isTransient, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queueChannel.QueueBind(queue.Name, key, exchange, false, nil)

	return queueChannel, queue, nil
}
