package latency

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
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
		w.WriteHeader(500)
		return
	}

	var shouldAddToQueue bool
	var delayReportID int32

	logrus.Print("this is method", r.Method)
	idAsString := r.FormValue("id")
	userDescription := r.FormValue("description")

	id, err := strconv.Atoi(idAsString)
	if err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(500)
		return
	}
	res, err := s.latencyQuerier.GetTripStatusAndOrderTimeDeliveryByOrderId(ctx, tx, int32(id))
	if err != nil {
		shouldAddToQueue = true
	}
	if res.TimeDelivery.Time.After(time.Now()) {
		w.WriteHeader(400)
		return
	}

	switch res.Status {
	case VENDOR_AT, PICKED, ASSIGNED:
		remainingTime, err := calculateRemainingTime()
		if err != nil {
			logrus.Errorf("could not get remaining time from api")
			w.WriteHeader(500)
			return
		}

		resp, err := json.Marshal(remainingTime)
		if err != nil {
			logrus.Errorf("could not marshal json")
			w.WriteHeader(500)
			return
		}

		w.Write(resp)
	case DELIVERED:
		shouldAddToQueue = true
	}
	isReportValid, err := s.latencyQuerier.CheckDelayReportOrderIDIsClosed(ctx, tx, pgtype.Int4{Int32: int32(id)})

	if isReportValid {
		delayReportID, err = s.addDelayReports(ctx, tx, id, userDescription)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		err = tx.Commit(ctx)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("Error:", err)
			return
		}

		if shouldAddToQueue {
			err = s.addToDelayQueue(ctx, delayReportID)
			if err != nil {
				w.WriteHeader(500)
				return
			}
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
	fmt.Println("Report data", description, orderID, delayReportID)
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
	resp, err := http.Get("https://run.mocky.io/v3/122c2796-5df4-461c-ab75-87c1192b17f7")
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	eta := response.Data.Eta
	return eta, nil
}
