package analyze

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// WriteJSONReport writes the report as JSON
func WriteJSONReport(report ImpactReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// PrintTextReport prints a human-readable report to console
func PrintTextReport(report ImpactReport) {
	fmt.Println()
	fmt.Println("CHAOS IMPACT REPORT")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("Chaos Applied: %s\n", report.ChaosDescription)
	fmt.Println()

	// Directly affected
	if len(report.DirectlyAffected) > 0 {
		fmt.Printf("DIRECTLY AFFECTED (%d):\n", len(report.DirectlyAffected))
		for _, comp := range report.DirectlyAffected {
			fmt.Printf("  ⚡ %s %s\n", comp.Method, comp.Path)
			fmt.Printf("     Latency: %.0fms → %.0fms (+%.0fms)\n",
				comp.Baseline.AvgLatencyMs,
				comp.Experiment.AvgLatencyMs,
				comp.AvgLatencyDelta)
		}
		fmt.Println()
	}

	// Critical impact
	if len(report.CriticalImpact) > 0 {
		fmt.Printf("CRITICAL IMPACT (%d):\n", len(report.CriticalImpact))
		for _, comp := range report.CriticalImpact {
			fmt.Printf("  ❌ %s %s\n", comp.Method, comp.Path)
			fmt.Printf("     Success Rate: %.1f%% → %.1f%% (%.1fpp)\n",
				comp.Baseline.SuccessRate,
				comp.Experiment.SuccessRate,
				comp.SuccessRateDelta)
			fmt.Printf("     Avg Latency: %.0fms → %.0fms (%+.0fms)\n",
				comp.Baseline.AvgLatencyMs,
				comp.Experiment.AvgLatencyMs,
				comp.AvgLatencyDelta)
			if comp.ErrorCountDelta != 0 {
				fmt.Printf("     Errors: %d → %d (%+d)\n",
					comp.Baseline.ErrorCount,
					comp.Experiment.ErrorCount,
					comp.ErrorCountDelta)
			}
		}
		fmt.Println()
	}

	// Major impact
	if len(report.MajorImpact) > 0 {
		fmt.Printf("MAJOR IMPACT (%d):\n", len(report.MajorImpact))
		for _, comp := range report.MajorImpact {
			fmt.Printf("  ⚠️  %s %s\n", comp.Method, comp.Path)
			fmt.Printf("     Success Rate: %.1f%% → %.1f%% (%.1fpp)\n",
				comp.Baseline.SuccessRate,
				comp.Experiment.SuccessRate,
				comp.SuccessRateDelta)
			fmt.Printf("     Avg Latency: %.0fms → %.0fms (%+.0fms)\n",
				comp.Baseline.AvgLatencyMs,
				comp.Experiment.AvgLatencyMs,
				comp.AvgLatencyDelta)
		}
		fmt.Println()
	}

	// Minor impact
	if len(report.MinorImpact) > 0 {
		fmt.Printf("MINOR IMPACT (%d):\n", len(report.MinorImpact))
		for _, comp := range report.MinorImpact {
			fmt.Printf("  ⚡ %s %s\n", comp.Method, comp.Path)
			fmt.Printf("     Success Rate: %.1f%% → %.1f%% (%.1fpp)\n",
				comp.Baseline.SuccessRate,
				comp.Experiment.SuccessRate,
				comp.SuccessRateDelta)
			fmt.Printf("     Avg Latency: %.0fms → %.0fms (%+.0fms)\n",
				comp.Baseline.AvgLatencyMs,
				comp.Experiment.AvgLatencyMs,
				comp.AvgLatencyDelta)
		}
		fmt.Println()
	}

	// Unaffected
	if len(report.Unaffected) > 0 {
		fmt.Printf("UNAFFECTED (%d):\n", len(report.Unaffected))
		for _, comp := range report.Unaffected {
			fmt.Printf("  %s %s\n", comp.Method, comp.Path)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("SUMMARY:")
	fmt.Printf("  Total Endpoints: %d\n", report.Summary.TotalEndpoints)
	fmt.Printf("  Hidden Dependencies: %d\n", report.Summary.HiddenDependencies)
	if report.Summary.CriticalImpact > 0 {
		fmt.Printf("  ❌ Critical: %d\n", report.Summary.CriticalImpact)
	}
	if report.Summary.MajorImpact > 0 {
		fmt.Printf("  ⚠️  Major: %d\n", report.Summary.MajorImpact)
	}
	if report.Summary.MinorImpact > 0 {
		fmt.Printf("  ⚡ Minor: %d\n", report.Summary.MinorImpact)
	}
	fmt.Printf("  Unaffected: %d\n", report.Summary.Unaffected)
	fmt.Println()

	// Recommendations
	if report.Summary.HiddenDependencies > 0 {
		fmt.Println("💡 RECOMMENDATIONS:")
		fmt.Println("  Hidden dependencies detected! Consider:")
		fmt.Println("  - Add circuit breakers to prevent cascade failures")
		fmt.Println("  - Implement retry logic with exponential backoff")
		fmt.Println("  - Cache auth tokens to reduce dependency calls")
		fmt.Println("  - Add timeouts to prevent hanging requests")
		fmt.Println()
	}
}
