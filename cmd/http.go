/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"chaos-cli/cmd/pkg/proxy"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("Starting chaos proxy with sample rules...")

	// 	rules := []proxy.ChaosRule{
	// 		{
	// 			Path:   "/delay",
	// 			Method: "GET",
	// 			Delay:  2 * time.Second,
	// 		},
	// 		{
	// 			Path:        "/fail",
	// 			Method:      "POST",
	// 			FailureRate: 0.5,
	// 			StatusCode:  503,
	// 			ErrorBody:   `{"error":"chaos injected"}`,
	// 		},
	// 	}

	// 	// Change this to your real backend
	// 	target := "http://localhost:8081"
	// 	port := 8080

	// 	p, err := proxy.NewChaosProxy(target, port, rules)
	// 	if err != nil {
	// 		fmt.Println("Error starting proxy:", err)
	// 		return
	// 	}
	// 	fmt.Printf("Chaos proxy running at :%d, forwarding to %s\n", port, target)
	// 	if err := p.Start(); err != nil {
	// 		fmt.Println("Proxy stopped with error:", err)
	// 	}
	// },
}

var (
	target      string
	port        int
	delay       time.Duration
	failureRate float64
	rulePath    string
	ruleMethod  string
)

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.AddCommand(httpProxyCmd)

	httpProxyCmd.Flags().StringVar(&target, "target", "http://localhost:3000", "Target server to proxy to")
	httpProxyCmd.Flags().IntVar(&port, "port", 8080, "Port to run the chaos proxy on")
	httpProxyCmd.Flags().DurationVar(&delay, "delay", 0, "Delay to inject")
	httpProxyCmd.Flags().Float64Var(&failureRate, "failure-rate", 0.0, "Failure rate (0.0 - 1.0)")
	httpProxyCmd.Flags().StringVar(&rulePath, "path", "/", "API path to match")
	httpProxyCmd.Flags().StringVar(&ruleMethod, "method", "", "HTTP method (GET, POST, etc.)")
}

var httpProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start HTTP chaos proxy",
	Run: func(cmd *cobra.Command, args []string) {

		rules := []proxy.ChaosRule{
			{
				Path:        rulePath,
				Method:      ruleMethod,
				Delay:       delay,
				FailureRate: failureRate,
				StatusCode:  503,
				ErrorBody:   "chaos injected via CLI",
			},
		}

		p, error := proxy.NewChaosProxy(target, port, rules)
		if error != nil {
			fmt.Println("Error starting proxy:", error)
			return
		}
		fmt.Printf("\n🚀 Chaos Proxy Running\n")
		fmt.Printf("▶ Forwarding:   %s\n", target)
		fmt.Printf("▶ Listening on: :%d\n", port)
		fmt.Printf("▶ Path Match:   %s\n", rulePath)
		fmt.Printf("▶ Method Match: %s\n", ruleMethod)
		fmt.Printf("▶ Delay:        %s\n", delay)
		fmt.Printf("▶ Fail Rate:    %.2f\n\n", failureRate)
		p.Start()
	},
}
