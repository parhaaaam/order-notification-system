package latency

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *servicer) GetVendorDelayReports(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		w.WriteHeader(500)
	}
	delayList, err := s.latencyQuerier.GetAllDelaysInLastWeek(ctx, tx)
	if err != nil {
		w.WriteHeader(500)
	}

	resp, err := json.Marshal(delayList)
	if err != nil {
		w.WriteHeader(500)
	}

	w.Write(resp)
}
