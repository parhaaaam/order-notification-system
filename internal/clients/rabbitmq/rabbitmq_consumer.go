package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	Queue    amqp091.Queue
	Messages <-chan amqp091.Delivery
	Channel  *amqp091.Channel
}

func NewConsumer(amqpChannel *amqp091.Channel, queueName string) (*Consumer, error) {
	q, err := amqpChannel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"Queue name": queueName,
		}).Errorf("could not create queue%s", err)
		return nil, err
	}

	if err := amqpChannel.Qos(1, 0, false); err != nil {
		return nil, err
	}

	msgs, err := amqpChannel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"Queue name":      q.Name,
			"Queue consumers": q.Consumers,
		}).Errorf("could not create consumer%s", err)
		return nil, err
	}

	return &Consumer{
		Queue:    q,
		Messages: msgs,
		Channel:  amqpChannel,
	}, nil
}

func (c *Consumer) ReadMessage() amqp091.Delivery {
	data := <-c.Messages
	return data
}
