/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/syedowais312/chaos-cli/pkg/proxy"
	"github.com/syedowais312/chaos-cli/pkg/utils"

	"github.com/spf13/cobra"
)

// If these are declared elsewhere in your repo (e.g. root flags), remove these declarations here.
var (
	duration time.Duration
	output   string
)

var (
	target      string
	port        int
	delay       time.Duration
	failureRate float64
	rulePath    string
	ruleMethod  string
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP commands for chaos proxy",
	Long: `HTTP subcommands for the chaos-cli. Examples:

  chaos-cli http proxy --target http://localhost:3000 --path /login --method POST --delay 2s --failure-rate 0.2
`,
}

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.AddCommand(httpProxyCmd)

	httpProxyCmd.Flags().StringVar(&target, "target", "http://localhost:3000", "Target server to proxy to")
	httpProxyCmd.Flags().IntVar(&port, "port", 8080, "Port to run the chaos proxy on")
	httpProxyCmd.Flags().DurationVar(&delay, "delay", 0, "Delay to inject")
	httpProxyCmd.Flags().Float64Var(&failureRate, "failure-rate", 0.0, "Failure rate (0.0 - 1.0)")
	httpProxyCmd.Flags().StringVar(&rulePath, "path", "/", "API path to match")
	httpProxyCmd.Flags().StringVar(&ruleMethod, "method", "", "HTTP method (GET, POST, etc.)")
	httpProxyCmd.Flags().DurationVar(&duration, "duration", 0, "Duration to run proxy (e.g., 60s). 0 means run until Ctrl+C")
    // Default metrics filename will be resolved into chaos-cli-test folder
    httpProxyCmd.Flags().StringVar(&output, "output", "baseline.ndjson", "NDJSON metrics filename (default: chaos-cli-test/baseline.ndjson)")
}

var httpProxyCmd = &cobra.Command{
    Use:   "proxy",
    Short: "Start HTTP chaos proxy",
    Run: func(cmd *cobra.Command, args []string) {
        // normalize method
        method := strings.ToUpper(strings.TrimSpace(ruleMethod))
        if method == "" {
            method = "" // empty means match any method (depends on your matching logic)
        }

		// Build rules from flags (you can extend to accept multiple rules)
		rules := []proxy.ChaosRule{
			{
				Path:        rulePath,
				Method:      method,
				Delay:       delay,
				FailureRate: failureRate,
				// It's probably better to make status and body flags; using defaults here
				StatusCode: 503,
				ErrorBody:  `{"error":"chaos injected"}`,
			},
		}

		p, err := proxy.NewChaosProxy(target, port, rules)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

        // Determine run label: baseline (record) vs experiment (test)
        runLabel := "record"
        runDetail := "baseline"
        if delay > 0 || failureRate > 0 {
            runLabel = "test"
            runDetail = "experiment"
        } else {
            // also infer from output filename when provided
            low := strings.ToLower(strings.TrimSpace(output))
            if strings.Contains(low, "experiment") {
                runLabel = "test"
                runDetail = "experiment"
            } else if strings.Contains(low, "baseline") {
                runLabel = "record"
                runDetail = "baseline"
            }
        }

        // Helpful startup logs so users know it is running
        fmt.Printf("Starting chaos proxy on :%d -> %s\n", port, target)
        fmt.Printf("Mode: %s (%s)\n", runLabel, runDetail)

        // If user did not specify --output, pick default by mode
        // record -> baseline.ndjson, test -> experiment.ndjson
        if !cmd.Flags().Changed("output") {
            if runLabel == "test" {
                output = "experiment.ndjson"
            } else {
                output = "baseline.ndjson"
            }
        }

		// If a duration is provided, run for that long and cancel
		if duration > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()

			go func() {
				// p.StartWithCtx is assumed to exist in your proxy package.
				// If it is named differently, update accordingly.
				if err := p.StartWithCtx(ctx); err != nil && err != http.ErrServerClosed {
					fmt.Println("proxy error:", err)
				}
			}()

			<-ctx.Done()
		} else {
			// Block until the proxy stops (Start handles signals internally)
			if err := p.Start(); err != nil && err != http.ErrServerClosed {
				fmt.Println("proxy error:", err)
			}
		}

		// At this point proxy was stopped; dump metrics
        fmt.Println("Proxy stopped; preparing metrics output...")
        
        if output == "" {
            fmt.Println("No output file provided; skipping metrics write.")
            return
        }
		outputPath, err := utils.ResolveOutputPath(output)
		if err != nil {
			fmt.Println("Failed to resolve output path:", err)
			return
		}
		fmt.Println("Resolved output path:", outputPath)

		if err := p.Metrics.WriteNDJSON(outputPath); err != nil {
			fmt.Println("Failed to write metrics:", err)
		} else {
			fmt.Println("Metrics written to", outputPath)
		}

	},
}
