package metrics

import "time"

type RequestMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int64     `json:"latency_ms"`
	ChaosApplied bool      `json:"chaos_applied"`
	ChaosType    string    `json:"chaos_type"`    // "delay", "failure", "none"
	BackendError bool      `json:"backend_error"` // true if backend returned 5xx or proxy detected error
}
