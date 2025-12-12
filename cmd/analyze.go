package cmd

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/spf13/cobra"
    "github.com/syedowais312/chaos-cli/pkg/analyze"
    "github.com/syedowais312/chaos-cli/pkg/utils"
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

    // Defaults to files inside chaos-cli-test folder
    analyzeCmd.Flags().StringVar(&baselineFile, "baseline", "baseline.ndjson", "Baseline metrics filename (default: chaos-cli-test/baseline.ndjson)")
    analyzeCmd.Flags().StringVar(&experimentFile, "experiment", "experiment.ndjson", "Experiment metrics filename (default: chaos-cli-test/experiment.ndjson)")
    analyzeCmd.Flags().StringVar(&outputFile, "output", "report.json", "Output report filename (default: chaos-cli-test/report.json)")
    analyzeCmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, brief, json, or both")
}

func runAnalyze(cmd *cobra.Command, args []string) {
    // Resolve paths into chaos-cli-test folder (creates it if missing)
    baselinePath, err := utils.ResolveOutputPath(baselineFile)
    if err != nil {
        log.Fatalf("Failed to resolve baseline path: %v", err)
    }
    experimentPath, err := utils.ResolveOutputPath(experimentFile)
    if err != nil {
        log.Fatalf("Failed to resolve experiment path: %v", err)
    }

    // Load baseline metrics
    baselineMetrics, err := loadMetrics(baselinePath)
    if err != nil {
        log.Fatalf("Failed to load baseline: %v", err)
    }

    // Load experiment metrics
    experimentMetrics, err := loadMetrics(experimentPath)
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
        outPath, err := utils.ResolveOutputPath(outputFile)
        if err != nil {
            log.Fatalf("Failed to resolve output path: %v", err)
        }
        if err := analyze.WriteJSONReport(report, outPath); err != nil {
            log.Fatalf("Failed to write JSON report: %v", err)
        }
        fmt.Printf("JSON report saved to %s\n", outPath)

    case "text":
        analyze.PrintTextReport(report)

    case "brief":
        analyze.PrintBriefReport(report)

    case "both":
        analyze.PrintTextReport(report)
        outPath, err := utils.ResolveOutputPath(outputFile)
        if err != nil {
            log.Fatalf("Failed to resolve output path: %v", err)
        }
        if err := analyze.WriteJSONReport(report, outPath); err != nil {
            log.Fatalf("Failed to write JSON report: %v", err)
        }
        fmt.Printf("JSON report saved to %s\n", outPath)

    default:
        log.Fatalf("Unknown format: %s (use 'text', 'brief', 'json', or 'both')", outputFormat)
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