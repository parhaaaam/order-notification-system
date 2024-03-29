package rest

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"order_notification_system/cmd/config"
	"order_notification_system/internal/clients/rabbitmq"
	"order_notification_system/internal/storage"
	"os"

	"order_notification_system/internal/latency"
	"order_notification_system/internal/storage/entities"
)

func ServeRest() error {
	ctx := context.Background()
	conf := config.Load()

	pool, err := storage.NewConnectionPool(ctx, conf)
	if err != nil {
		return err
	}

	conn, err := amqp.Dial(conf.RabbitMQClient.ConnectionURL)
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Connection URL": conf.RabbitMQClient.ConnectionURL,
		}).Errorf("could not connect to rabbitmq%s", err)
		return err
	}

	amqpChannel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer amqpChannel.Close()

	delayReportConsumer, err := rabbitmq.NewConsumer(amqpChannel, "delayed_orders")
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Amqp Channel": amqpChannel,
		}).Errorf("could not create consumer%s", err)
		return err
	}

	latencyQuerier := entities.New()
	servicer := latency.NewServicer(pool, latencyQuerier, amqpChannel, delayReportConsumer)

	mux := http.NewServeMux()

	mux.HandleFunc("/get_vendor_delay_reports", servicer.GetVendorDelayReports)
	mux.HandleFunc("/order_delay_notification", servicer.GetOrderDelayNotification)
	mux.HandleFunc("/agent_approval", servicer.AgentApproval)

	err = http.ListenAndServe(":8080", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
		os.Exit(1)
	}

	return err
}
