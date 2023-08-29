package latency

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"strconv"

	"order_notification_system/internal/storage/entities"
)

func (s *servicer) AgentApproval(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	agentIDAsString := r.URL.Query().Get("agent_id")
	agentID, err := strconv.Atoi(agentIDAsString)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Reading message queue")
	msg := s.delayedOrderQueue.ReadMessage()
	var delayReportID int
	err = json.Unmarshal(msg.Body, &delayReportID)
	fmt.Println("New report received", delayReportID)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error:", err)
		return
	}

	agentIDToPgtypeInt := pgtype.Int4{
		Int32: int32(agentID),
		Valid: true,
	}
	s.latencyQuerier.AssignOrderToAgent(ctx, tx, entities.AssignOrderToAgentParams{
		AgentID: agentIDToPgtypeInt,
		ID:      int32(delayReportID),
	})

	err = tx.Commit(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error:", err)
		return
	}
}
