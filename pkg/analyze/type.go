package analyze

import "time"

// RequestMetric mirrors the metrics emitted by the proxy (NDJSON)
type RequestMetric struct {
    Timestamp    time.Time `json:"timestamp"`
    Method       string    `json:"method"`
    Path         string    `json:"path"`
    StatusCode   int       `json:"status_code"`
    LatencyMs    int64     `json:"latency_ms"`
    ChaosApplied bool      `json:"chaos_applied"`
    ChaosType    string    `json:"chaos_type"`
    BackendError bool      `json:"backend_error"`
}

// EndpointStats aggregates metrics for a single method+path
type EndpointStats struct {
    Method       string  `json:"method"`
    Path         string  `json:"path"`
    RequestCount int64   `json:"request_count"`
    SuccessRate  float64 `json:"success_rate"`
    AvgLatencyMs float64 `json:"avg_latency_ms"`
    P50LatencyMs float64 `json:"p50_latency_ms"`
    P95LatencyMs float64 `json:"p95_latency_ms"`
    P99LatencyMs float64 `json:"p99_latency_ms"`
    ErrorCount   int64   `json:"error_count"`
    ChaosApplied bool    `json:"chaos_applied"`
}

// EndpointComparison compares baseline vs experiment stats
type EndpointComparison struct {
    Method           string        `json:"method"`
    Path             string        `json:"path"`
    Baseline         EndpointStats `json:"baseline"`
    Experiment       EndpointStats `json:"experiment"`
    SuccessRateDelta float64       `json:"success_rate_delta"`
    AvgLatencyDelta  float64       `json:"avg_latency_delta"`
    ErrorCountDelta  int64         `json:"error_count_delta"`
    ImpactLevel      string        `json:"impact_level"`
}

// ReportSummary summarizes impact categories
type ReportSummary struct {
    TotalEndpoints     int `json:"total_endpoints"`
    DirectlyAffected   int `json:"directly_affected"`
    CriticalImpact     int `json:"critical_impact"`
    MajorImpact        int `json:"major_impact"`
    MinorImpact        int `json:"minor_impact"`
    Unaffected         int `json:"unaffected"`
    HiddenDependencies int `json:"hidden_dependencies"`
}

// ImpactReport is the main output data structure
type ImpactReport struct {
    ChaosDescription string              `json:"chaos_description"`
    DirectlyAffected []EndpointComparison `json:"directly_affected"`
    CriticalImpact   []EndpointComparison `json:"critical_impact"`
    MajorImpact      []EndpointComparison `json:"major_impact"`
    MinorImpact      []EndpointComparison `json:"minor_impact"`
    Unaffected       []EndpointComparison `json:"unaffected"`
    Summary          ReportSummary        `json:"summary"`
}