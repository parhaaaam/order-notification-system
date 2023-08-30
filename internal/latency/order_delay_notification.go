package latency

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"order_notification_system/cmd/config"
	"strconv"
	"time"

	"order_notification_system/internal/storage/entities"
)

type Response struct {
	Status bool `json:"status"`
	Data   struct {
		Eta int `json:"eta"`
	} `json:"data"`
}

func (s *servicer) GetOrderDelayNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("could not create a pgx pool %s", err)
		w.WriteHeader(500)
		return
	}

	var shouldAddToQueue bool
	var delayReportID int32

	orderIDAsString := r.FormValue("id")
	userDescription := r.FormValue("description")

	orderID, err := strconv.Atoi(orderIDAsString)
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Order ID": orderIDAsString,
		}).Errorf("could not convert order id %s", err)
		w.WriteHeader(500)
		return
	}

	res, err := s.latencyQuerier.GetTripStatusAndOrderTimeDeliveryByOrderId(ctx, tx, int32(orderID))
	if err != nil {
		shouldAddToQueue = true
	}

	if res.TimeDelivery.Time.After(time.Now()) {
		log.WithContext(ctx).WithFields(log.Fields{
			"Order ID":      orderID,
			"Delivery time": res.TimeDelivery,
		}).Errorf("time delivery from order is has not arriven %s", err)
		w.WriteHeader(400)
		return
	}

	switch res.Status {
	case AT_VENDOR, PICKED, ASSIGNED:
		remainingTime, err := calculateRemainingTime()
		if err != nil {
			log.WithContext(ctx).WithFields(log.Fields{
				"Status": res.Status,
			}).Errorf("could not get remaining time from remaining time calculator api %s", err)
			w.WriteHeader(500)
			return
		}

		resp, err := json.Marshal(remainingTime)
		if err != nil {
			log.WithContext(ctx).WithFields(log.Fields{
				"Remaining time": remainingTime,
			}).Errorf("could not convert remaining time %s", err)
			w.WriteHeader(500)
			return
		}

		w.Write(resp)
	case DELIVERED:
		shouldAddToQueue = true
	}

	isReportValid, err := s.latencyQuerier.CheckDelayReportOrderIDIsClosed(ctx, tx, pgtype.Int4{Int32: int32(orderID)})
	if isReportValid {
		delayReportID, err = s.addDelayReports(ctx, tx, orderID, userDescription)
		if err != nil {
			log.WithContext(ctx).WithFields(log.Fields{
				"Order ID":         orderID,
				"User description": userDescription,
			}).Errorf("could not add delay report in db %s", err)
			w.WriteHeader(500)
			return
		}

		if shouldAddToQueue {
			err = s.addToDelayQueue(ctx, delayReportID)
			if err != nil {
				log.WithContext(ctx).WithFields(log.Fields{
					"Delay report ID": delayReportID,
				}).Errorf("could not add delay report in queue %s", err)
				w.WriteHeader(500)
				return
			}
		}

		err = tx.Commit(ctx)
		if err != nil {
			log.WithContext(ctx).WithFields(log.Fields{
				"Order ID": orderID,
			}).Errorf("failed to commit transaction %s", err)
			w.WriteHeader(500)
			return
		}
	}
}

func (s *servicer) addDelayReports(ctx context.Context, tx pgx.Tx, id int,
	userDescription string) (int32, error) {
	description := pgtype.Text{
		String: userDescription,
		Valid:  true,
	}
	orderID := pgtype.Int4{
		Int32: int32(id),
		Valid: true,
	}

	delayReportID, err := s.latencyQuerier.AddDelayReports(ctx, tx, entities.AddDelayReportsParams{
		Description: description,
		OrderID:     orderID,
	})
	if err != nil {
		return 0, err
	}

	return delayReportID, nil
}

func (s *servicer) addToDelayQueue(ctx context.Context, delayReportID int32) error {
	body, err := json.Marshal(delayReportID)
	if err != nil {
		return err
	}

	s.amqpChannel.PublishWithContext(
		ctx,
		"",
		"delayed_orders",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)

	return nil
}

func calculateRemainingTime() (int, error) {
	conf := config.Load()
	resp, err := http.Get(conf.CalculatorAPI.Domain + conf.CalculatorAPI.Address)
	if err != nil {
		log.Errorf("could not get remaining time from calculator api %s", err)
		return 0, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("could not read response body from calculator api %s", err)
		return 0, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Errorf("could not convert response body to string %s", err)
		return 0, err
	}

	eta := response.Data.Eta
	return eta, nil
}
