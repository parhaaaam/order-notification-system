package rabbitmq

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestConsumer_ReadMessage(t *testing.T) {
	amqpChannel := &amqp091.Channel{}

	deliveryChannel := make(chan amqp091.Delivery)

	consumer := &Consumer{
		Messages: deliveryChannel,
		Channel:  amqpChannel,
	}

	expectedDelivery := amqp091.Delivery{
		Body: []byte("Hello, World!"),
	}

	go func() {
		deliveryChannel <- expectedDelivery
	}()

	actualDelivery := consumer.ReadMessage()

	assert.Equal(t, expectedDelivery, actualDelivery)
}
