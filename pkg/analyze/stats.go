package analyze

import (
	"sort"
)

func CalculateStats(metrics []RequestMetric) EndpointStats {
	if len(metrics) == 0 {
		return EndpointStats{}
	}
	stats := EndpointStats{
		Method:       metrics[0].Method,
		Path:         metrics[0].Path,
		RequestCount: int64(len(metrics)),
	}

	successCount := 0
	var latencies []int64
	chaosApplied := false

	for _, m := range metrics {
		if m.StatusCode >= 200 && m.StatusCode < 300 {
			successCount++
		} else {
			stats.ErrorCount++
		}

		latencies = append(latencies, m.LatencyMs)
		if m.ChaosApplied {
			chaosApplied = true
		}
	}

	stats.SuccessRate = float64(successCount) / float64(len(metrics)) * 100
	stats.ChaosApplied = chaosApplied

	var sum int64

	for _, latency := range latencies {
		sum += latency
	}
	stats.AvgLatencyMs = float64(sum) / float64(len(latencies))

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	stats.P50LatencyMs = float64(percentile(latencies, 0.50))
	stats.P95LatencyMs = float64(percentile(latencies, 0.95))
	stats.P99LatencyMs = float64(percentile(latencies, 0.99))

	return stats

}
func percentile(sortedLatencies []int64, p float64) int64 {
	if len(sortedLatencies) == 0 {
		return 0
	}

	index := int(float64(len(sortedLatencies)-1) * p)
	return sortedLatencies[index]
}

func GroupMetricsByEndpoint(metrics []RequestMetric) map[string][]RequestMetric {
	grouped := make(map[string][]RequestMetric)

	for _, m := range metrics {
		key := m.Method + ":" + m.Path
		grouped[key] = append(grouped[key], m)
	}

	return grouped
}
