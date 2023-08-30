package latency

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"order_notification_system/internal/storage/entities"
)

func (s *servicer) AgentApproval(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("could not create a pgx pool %s", err)
		w.WriteHeader(500)
		return
	}

	agentIDAsString := r.URL.Query().Get("agent_id")
	agentID, err := strconv.Atoi(agentIDAsString)
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Agent ID": agentIDAsString,
		}).Errorf("could not convert agent id %s", err)
		w.WriteHeader(500)
		return
	}

	err = s.assignOrderToAgentFromQueue(ctx, tx, agentID)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Agent ID": agentID,
		}).Errorf("failed to commit transaction %s", err)
		w.WriteHeader(500)
		return
	}
}

func (s *servicer) assignOrderToAgentFromQueue(ctx context.Context, tx pgx.Tx, agentID int) error {
	var delayReportID int
	msg := s.delayedOrderQueue.ReadMessage()
	err := json.Unmarshal(msg.Body, &delayReportID)
	if err != nil {
		log.WithContext(ctx).WithFields(log.Fields{
			"Queue message body": msg.Body,
			"Delay report ID":    delayReportID,
		}).Errorf("could not convert message body %s", err)
		return err
	}

	agentIDToPgtypeInt := pgtype.Int4{
		Int32: int32(agentID),
		Valid: true,
	}
	s.latencyQuerier.AssignOrderToAgent(ctx, tx, entities.AssignOrderToAgentParams{
		AgentID: agentIDToPgtypeInt,
		ID:      int32(delayReportID),
	})

	return nil
}
