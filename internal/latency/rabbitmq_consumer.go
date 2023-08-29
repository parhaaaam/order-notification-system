package latency

import "github.com/rabbitmq/amqp091-go"

type consumer struct {
	Queue    amqp091.Queue
	Messages <-chan amqp091.Delivery
	Channel  *amqp091.Channel
}

func NewConsumer(amqpChannel *amqp091.Channel, queueName string) (*consumer, error) {
	q, err := amqpChannel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	if err := amqpChannel.Qos(1, 0, false); err != nil {
		return nil, err
	}

	msgs, err := amqpChannel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		return nil, err
	}

	return &consumer{
		Queue:    q,
		Messages: msgs,
		Channel:  amqpChannel,
	}, nil
}

func (c *consumer) ReadMessage() amqp091.Delivery {
	data := <-c.Messages
	return data
}
