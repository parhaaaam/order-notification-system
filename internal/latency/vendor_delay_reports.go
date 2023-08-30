package latency

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *servicer) GetVendorDelayReports(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("could not create a pgx pool %s", err)
		w.WriteHeader(500)
		return
	}

	delayList, err := s.latencyQuerier.GetAllDelaysInLastWeek(ctx, tx)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to get delays in last week from db %s", err)
		w.WriteHeader(500)
	}

	resp, err := json.Marshal(delayList)
	if err != nil {
		log.WithContext(ctx).Errorf("could not convert delays into json %s", err)
		w.WriteHeader(500)
	}

	w.Write(resp)
}
