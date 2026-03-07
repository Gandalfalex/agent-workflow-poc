package httpapi

import (
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/store"
)

func (h *API) GetFlowHealth(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}

	stats, err := h.store.GetFlowHealthStats(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "flow_health_failed", "failed to get flow health stats")
		return
	}

	writeJSON(w, http.StatusOK, mapFlowHealthStats(stats))
}

func mapFlowHealthStats(s store.FlowHealthStats) flowHealthStatsResponse {
	wipStats := make([]FlowHealthWipStat, 0, len(s.WipStats))
	for _, ws := range s.WipStats {
		wipStats = append(wipStats, FlowHealthWipStat{
			StateId:        openapi_types.UUID(ws.StateID),
			StateName:      ws.StateName,
			WipLimit:       ws.WipLimit,
			WipEnforcement: ws.WipEnforcement,
			TicketCount:    ws.TicketCount,
			MaxAgeHours:    float32(ws.MaxAgeHours),
			AvgAgeHours:    float32(ws.AvgAgeHours),
		})
	}

	throughput := make([]ThroughputDay, 0, len(s.Throughput))
	for _, t := range s.Throughput {
		throughput = append(throughput, ThroughputDay{
			Day:   openapi_types.Date{Time: t.Day},
			Count: t.Count,
		})
	}

	return flowHealthStatsResponse{
		WipStats: wipStats,
		CycleTime: CycleTimePercentiles{
			P50Hours:    float32(s.CycleTime.P50Hours),
			P75Hours:    float32(s.CycleTime.P75Hours),
			P95Hours:    float32(s.CycleTime.P95Hours),
			SampleCount: s.CycleTime.SampleCount,
		},
		Throughput: throughput,
	}
}
