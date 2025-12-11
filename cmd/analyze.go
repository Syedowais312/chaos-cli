package cmd

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/spf13/cobra"
    "github.com/syedowais312/chaos-cli/pkg/analyze"
)

var analyzeCmd = &cobra.Command{
    Use:   "analyze",
    Short: "Analyze chaos experiment results",
    Long:  `Compare baseline and experiment metrics to identify dependencies`,
    Run:   runAnalyze,
}

var (
    baselineFile   string
    experimentFile string
    outputFile     string
    outputFormat   string
)

func init() {
    httpCmd.AddCommand(analyzeCmd)

    analyzeCmd.Flags().StringVar(&baselineFile, "baseline", "", "Baseline metrics file (required)")
    analyzeCmd.Flags().StringVar(&experimentFile, "experiment", "", "Experiment metrics file (required)")
    analyzeCmd.Flags().StringVar(&outputFile, "output", "report.json", "Output report file")
    analyzeCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json, or both")

    analyzeCmd.MarkFlagRequired("baseline")
    analyzeCmd.MarkFlagRequired("experiment")
}

func runAnalyze(cmd *cobra.Command, args []string) {
    // Load baseline metrics
    baselineMetrics, err := loadMetrics(baselineFile)
    if err != nil {
        log.Fatalf("Failed to load baseline: %v", err)
    }

    // Load experiment metrics
    experimentMetrics, err := loadMetrics(experimentFile)
    if err != nil {
        log.Fatalf("Failed to load experiment: %v", err)
    }

    fmt.Printf("Loaded %d baseline metrics and %d experiment metrics\n",
        len(baselineMetrics), len(experimentMetrics))

    // Extract chaos description
    chaosDesc := analyze.GetChaosDescription(experimentMetrics)

    // Generate report
    report := analyze.GenerateImpactReport(baselineMetrics, experimentMetrics, chaosDesc)

    // Output based on format
    switch outputFormat {
    case "json":
        if err := analyze.WriteJSONReport(report, outputFile); err != nil {
            log.Fatalf("Failed to write JSON report: %v", err)
        }
        fmt.Printf("JSON report saved to %s\n", outputFile)

    case "text":
        analyze.PrintTextReport(report)

    case "both":
        analyze.PrintTextReport(report)
        if err := analyze.WriteJSONReport(report, outputFile); err != nil {
            log.Fatalf("Failed to write JSON report: %v", err)
        }
        fmt.Printf("JSON report saved to %s\n", outputFile)

    default:
        log.Fatalf("Unknown format: %s (use 'text', 'json', or 'both')", outputFormat)
    }
}

// loadMetrics loads metrics from an NDJSON file
func loadMetrics(filename string) ([]analyze.RequestMetric, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var metrics []analyze.RequestMetric
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        var metric analyze.RequestMetric
        if err := json.Unmarshal(scanner.Bytes(), &metric); err != nil {
            return nil, fmt.Errorf("failed to parse line: %w", err)
        }
        metrics = append(metrics, metric)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return metrics, nil
}