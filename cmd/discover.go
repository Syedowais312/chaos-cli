package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/syedowais312/chaos-cli/pkg/discover"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover API endpoints by observing traffic",
	Long:  `Runs a reverse proxy that records all unique endpoints it sees.`,
	Run:   runDiscover,
}

var (
	discoverTarget   string
	discoverPort     string
	discoverDuration int
	discoverOutput   string
)

func init() {
    // Register discover as a top-level command for simplicity
    rootCmd.AddCommand(discoverCmd)

	discoverCmd.Flags().StringVar(&discoverTarget, "target", "", "Backend target URL (required)")
	discoverCmd.Flags().StringVar(&discoverPort, "port", "8080", "Proxy listen port")
	discoverCmd.Flags().IntVar(&discoverDuration, "duration", 0, "Auto-stop after N seconds (0 = manual)")
	discoverCmd.Flags().StringVar(&discoverOutput, "output", "endpoints.json", "Output file for discovered endpoints")

    discoverCmd.MarkFlagRequired("target")
}

func runDiscover(cmd *cobra.Command, args []string) {
	targetURL, err := url.Parse(discoverTarget)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	collector := discover.NewEndpointCollector()

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		collector.RecordEndpoint(r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    ":" + discoverPort,
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting endpoint discovery on :%s -> %s\n", discoverPort, discoverTarget)

		if discoverDuration > 0 {
			fmt.Printf("Will auto-stop after %d seconds\n", discoverDuration)
		} else {
			fmt.Println("Press Ctrl+C to stop and save discovered endpoints")
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if discoverDuration > 0 {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal")
		case <-time.After(time.Duration(discoverDuration) * time.Second):
			fmt.Println("\nDuration elapsed")
		}
	} else {
		<-sigChan
		fmt.Println("\nReceived interrupt signal")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Save endpoints
	fmt.Printf("Saving discovered endpoints to %s...\n", discoverOutput)
	if err := collector.WriteToFile(discoverOutput); err != nil {
		log.Fatalf("Failed to write endpoints: %v", err)
	}

	endpoints := collector.GetEndpoints()
	fmt.Printf("✅ Discovered %d unique endpoints\n", len(endpoints))
	for _, ep := range endpoints {
		fmt.Printf("   %s %s\n", ep.Method, ep.Path)
	}
}
