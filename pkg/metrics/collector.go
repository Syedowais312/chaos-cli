package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// MetricsCollector is thread-safe and stores metrics in memory
type MetricsCollector struct {
	mu      sync.Mutex
	metrics []RequestMetric
}

// New creates collector
func New() *MetricsCollector {
	return &MetricsCollector{
		metrics: make([]RequestMetric, 0, 1024),
	}
}

// RecordRequest appends a metric (concurrent-safe)
func (c *MetricsCollector) RecordRequest(m RequestMetric) {
	c.mu.Lock()
	c.metrics = append(c.metrics, m)
	c.mu.Unlock()
}

// GetAll returns a snapshot copy of collected metrics
func (c *MetricsCollector) GetAll() []RequestMetric {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]RequestMetric, len(c.metrics))
	copy(cp, c.metrics)
	return cp
}

// Clear removes stored metrics
func (c *MetricsCollector) Clear() {
	c.mu.Lock()
	c.metrics = c.metrics[:0]
	c.mu.Unlock()
}

// WriteNDJSON writes metrics to a file (newline-delimited JSON)
func (c *MetricsCollector) WriteNDJSON(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)

	for _, m := range c.metrics {
		if err := enc.Encode(m); err != nil {
			w.Flush()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

// AppendNDJSON appends metrics to a file (useful to stream metrics)
func (c *MetricsCollector) AppendNDJSON(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)

	for _, m := range c.metrics {
		if err := enc.Encode(m); err != nil {
			w.Flush()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func (p *Proxy) Start(duration time.Duration) {
	if duration > 0 {
		// ...existing duration logic...
	} else {
		if err := p.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Println("proxy error:", err)
		}
	}
}
