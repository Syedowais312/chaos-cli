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
	httpProxyCmd.Flags().StringVar(&output, "output", "", "File to write NDJSON metrics on shutdown")
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
			go func() {
				if err := p.Start(); err != nil && err != http.ErrServerClosed {
					fmt.Println("proxy error:", err)
				}
			}()
			// Wait until process is interrupted (Start handles signals)
			select {}
		}

		// At this point proxy was stopped; dump metrics
		if output != "" {
			if err := p.Metrics.WriteNDJSON(output); err != nil {
				fmt.Println("failed to write metrics:", err)
			} else {
				fmt.Println("metrics written to", output)
			}
		} else {
			fmt.Println("no output file provided; metrics kept in memory")
		}
	},
}
