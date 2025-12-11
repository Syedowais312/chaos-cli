package analyze

import (
    "fmt"
)

// CompareEndpoints compares baseline and experiment stats
func CompareEndpoints(baseline, experiment EndpointStats) EndpointComparison {
    comparison := EndpointComparison{
        Method:     baseline.Method,
        Path:       baseline.Path,
        Baseline:   baseline,
        Experiment: experiment,
    }

    // Calculate deltas
    comparison.SuccessRateDelta = experiment.SuccessRate - baseline.SuccessRate
    comparison.AvgLatencyDelta = experiment.AvgLatencyMs - baseline.AvgLatencyMs
    comparison.ErrorCountDelta = experiment.ErrorCount - baseline.ErrorCount

    // Classify impact level
    comparison.ImpactLevel = classifyImpact(comparison)

    return comparison
}

// classifyImpact determines the severity of impact
func classifyImpact(comp EndpointComparison) string {
    // If chaos was directly applied to this endpoint
    if comp.Experiment.ChaosApplied {
        return "directly_affected"
    }

    // Critical: success rate drops >20% OR latency increases >1000ms
    if comp.SuccessRateDelta < -20 || comp.AvgLatencyDelta > 1000 {
        return "critical"
    }

    // Major: success rate drops >10% OR latency increases >500ms
    if comp.SuccessRateDelta < -10 || comp.AvgLatencyDelta > 500 {
        return "major"
    }

    // Minor: success rate drops >5% OR latency increases >150ms
    if comp.SuccessRateDelta < -5 || comp.AvgLatencyDelta > 150 {
        return "minor"
    }

    return "none"
}

// GenerateImpactReport creates a full analysis report
func GenerateImpactReport(baselineMetrics, experimentMetrics []RequestMetric, chaosDescription string) ImpactReport {
    report := ImpactReport{
        ChaosDescription: chaosDescription,
    }

    // Group metrics by endpoint
    baselineGrouped := GroupMetricsByEndpoint(baselineMetrics)
    experimentGrouped := GroupMetricsByEndpoint(experimentMetrics)

    // Find all unique endpoints
    allEndpoints := make(map[string]bool)
    for key := range baselineGrouped {
        allEndpoints[key] = true
    }
    for key := range experimentGrouped {
        allEndpoints[key] = true
    }

    // Compare each endpoint
    for key := range allEndpoints {
        baselineStats := CalculateStats(baselineGrouped[key])
        experimentStats := CalculateStats(experimentGrouped[key])

        // Skip if endpoint doesn't exist in both runs
        if len(baselineGrouped[key]) == 0 || len(experimentGrouped[key]) == 0 {
            continue
        }

        comparison := CompareEndpoints(baselineStats, experimentStats)

        // Categorize by impact level
        switch comparison.ImpactLevel {
        case "directly_affected":
            report.DirectlyAffected = append(report.DirectlyAffected, comparison)
        case "critical":
            report.CriticalImpact = append(report.CriticalImpact, comparison)
        case "major":
            report.MajorImpact = append(report.MajorImpact, comparison)
        case "minor":
            report.MinorImpact = append(report.MinorImpact, comparison)
        case "none":
            report.Unaffected = append(report.Unaffected, comparison)
        }
    }

    // Generate summary
    report.Summary = ReportSummary{
        TotalEndpoints:     len(allEndpoints),
        DirectlyAffected:   len(report.DirectlyAffected),
        CriticalImpact:     len(report.CriticalImpact),
        MajorImpact:        len(report.MajorImpact),
        MinorImpact:        len(report.MinorImpact),
        Unaffected:         len(report.Unaffected),
        HiddenDependencies: len(report.CriticalImpact) + len(report.MajorImpact) + len(report.MinorImpact),
    }

    return report
}

// GetChaosDescription extracts chaos description from experiment metrics
func GetChaosDescription(metrics []RequestMetric) string {
    for _, m := range metrics {
        if m.ChaosApplied {
            return fmt.Sprintf("%s on %s %s", m.ChaosType, m.Method, m.Path)
        }
    }
    return "unknown chaos"
}