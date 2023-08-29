package latency

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	Storage "order_notification_system/internal/storage/entities"
)

type servicer struct {
	pool *pgxpool.Pool

	latencyQuerier Storage.Querier

	amqpChannel       *amqp091.Channel
	delayedOrderQueue *consumer
}

func NewServicer(pool *pgxpool.Pool, latencyQuerier Storage.Querier,
	amqpChannel *amqp091.Channel, delayedOrderQueue *consumer) *servicer {
	return &servicer{
		pool:              pool,
		latencyQuerier:    latencyQuerier,
		amqpChannel:       amqpChannel,
		delayedOrderQueue: delayedOrderQueue,
	}
}
