package latency

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"order_notification_system/internal/clients/rabbitmq"
	Storage "order_notification_system/internal/storage/entities"
)

type servicer struct {
	pool *pgxpool.Pool

	latencyQuerier Storage.Querier

	amqpChannel       *amqp091.Channel
	delayedOrderQueue *rabbitmq.Consumer
}

func NewServicer(pool *pgxpool.Pool, latencyQuerier Storage.Querier,
	amqpChannel *amqp091.Channel, delayedOrderQueue *rabbitmq.Consumer) *servicer {
	return &servicer{
		pool:              pool,
		latencyQuerier:    latencyQuerier,
		amqpChannel:       amqpChannel,
		delayedOrderQueue: delayedOrderQueue,
	}
}
