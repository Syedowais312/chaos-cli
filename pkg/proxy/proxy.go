package proxy

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/syedowais312/chaos-cli/pkg/metrics"
)

// type ChaosRule struct {
// 	Path        string
// 	Method      string
// 	Delay       time.Duration
// 	FailureRate float64
// 	StatusCode  int
// 	ErrorBody   string
// }

type ChaosProxy struct {
	TargetURL *url.URL
	Port      int
	Rules     []ChaosRule
	proxy     *httputil.ReverseProxy
	Metrics   *metrics.MetricsCollector
	server    *http.Server
}

// NewChaosProxy creates a configured proxy
func NewChaosProxy(target string, port int, rules []ChaosRule) (*ChaosProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	rp := httputil.NewSingleHostReverseProxy(u)

	cp := &ChaosProxy{
		TargetURL: u,
		Port:      port,
		Rules:     rules,
		proxy:     rp,
		Metrics:   metrics.New(),
	}
	// keep default director, but you can override if needed
	return cp, nil
}

func (cp *ChaosProxy) StartWithCtx(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", cp)

	addr := fmt.Sprintf(":%d", cp.Port)
	cp.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// run server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- cp.server.ListenAndServe()
	}()

	// listen for ctx done or signal
	select {
	case <-ctx.Done():
		// shutdown triggered by caller
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return cp.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		// server returned an error
		return err
	}
}

// Start runs until SIGINT or SIGTERM and then dumps metrics to outputFile if set
func (cp *ChaosProxy) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// OS signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		// wait for signal
		<-sigCh
		cancel()
	}()

	return cp.StartWithCtx(ctx)
}

// findMatchingRule finds the first matching rule for path+method
func (cp *ChaosProxy) findMatchingRule(path, method string) *ChaosRule {
	for i := range cp.Rules {
		r := &cp.Rules[i]
		if r.Path != "" && r.Path != path {
			continue
		}
		if r.Method != "" && r.Method != method {
			continue
		}
		return r
	}
	return nil
}

func (cp *ChaosProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	chaosApplied := false
	chaosType := "none"
	backendErr := false

	// Check rule
	rule := cp.findMatchingRule(r.URL.Path, r.Method)
	if rule != nil {
		// Delay
		if rule.Delay > 0 {
			chaosApplied = true
			chaosType = "delay"
			time.Sleep(rule.Delay)
		}
		// Random fail
		if rule.FailureRate > 0 && rand.Float64() < rule.FailureRate {
			chaosApplied = true
			chaosType = "failure"
			status := rule.StatusCode
			if status == 0 {
				status = http.StatusServiceUnavailable
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			if rule.ErrorBody != "" {
				_, _ = w.Write([]byte(rule.ErrorBody))
			} else {
				_, _ = w.Write([]byte(`{"error":"chaos injected"}`))
			}
			lat := time.Since(start).Milliseconds()
			cp.Metrics.RecordRequest(metrics.RequestMetric{
				Timestamp:    time.Now(),
				Method:       r.Method,
				Path:         r.URL.Path,
				StatusCode:   status,
				LatencyMs:    lat,
				ChaosApplied: true,
				ChaosType:    chaosType,
				BackendError: false,
			})
			return
		}
	}

	// Wrap ResponseWriter to capture status
	rec := NewStatusRecorder(w)
	cp.proxy.ServeHTTP(rec, r)

	// Determine backend error (5xx)
	if rec.StatusCode >= 500 {
		backendErr = true
	}

	lat := time.Since(start).Milliseconds()
	cp.Metrics.RecordRequest(metrics.RequestMetric{
		Timestamp:    time.Now(),
		Method:       r.Method,
		Path:         r.URL.Path,
		StatusCode:   rec.StatusCode,
		LatencyMs:    lat,
		ChaosApplied: chaosApplied,
		ChaosType:    chaosType,
		BackendError: backendErr,
	})
}
