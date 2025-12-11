package discover

import (
    "encoding/json"
    "os"
    "sync"
    "time"
)

type EndpointCollector struct {
    endpoints map[string]Endpoint
    mu        sync.RWMutex
}

func NewEndpointCollector() *EndpointCollector {
    return &EndpointCollector{
        endpoints: make(map[string]Endpoint),
    }
}

func (c *EndpointCollector) RecordEndpoint(method, path string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    key := method + ":" + path
    if _, exists := c.endpoints[key]; !exists {
        c.endpoints[key] = Endpoint{
            Method: method,
            Path:   normalizePath(path),
        }
    }
}

// GetEndpoints returns a snapshot slice of unique endpoints
func (c *EndpointCollector) GetEndpoints() []Endpoint {
    c.mu.RLock()
    defer c.mu.RUnlock()
    endpoints := make([]Endpoint, 0, len(c.endpoints))
    for _, ep := range c.endpoints {
        endpoints = append(endpoints, ep)
    }
    return endpoints
}

// Write writes discovered endpoints to a file in JSON format
func (c *EndpointCollector) Write(filename string) error {
    return c.WriteToFile(filename)
}

// WriteToFile writes discovered endpoints to a file in JSON format
func (c *EndpointCollector) WriteToFile(filename string) error {
    endpoints := c.GetEndpoints()

    list := EndpointList{
        Endpoints: endpoints,
        Source:    "passive",
        Timestamp: time.Now().Format(time.RFC3339),
    }

    data, err := json.MarshalIndent(list, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(filename, data, 0644)
}

func normalizePath(path string) string {
    // for now returning the path as it is
    // in future will add smart normalizePath for IDs, UUIDs, etc.
    return path
}