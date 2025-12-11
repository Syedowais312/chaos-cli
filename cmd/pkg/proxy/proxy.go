package proxy

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ChaosProxy struct {
	TargetURL *url.URL
	Port      int
	Rules     []ChaosRule
	Proxy     *httputil.ReverseProxy
	server    *http.Server
}

func NewChaosProxy(target string, port int, rules []ChaosRule) (*ChaosProxy, error) {
	parsedURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %v", err)
	}

	rp := httputil.NewSingleHostReverseProxy(parsedURL)

	cp := &ChaosProxy{
		TargetURL: parsedURL,
		Port:      port,
		Rules:     rules,
		Proxy:     rp,
	}
	return cp, nil
}

func (cp *ChaosProxy) findRule(path, method string) *ChaosRule {
	for i := range cp.Rules {
		rule := &cp.Rules[i]

		if rule.Path != "" && rule.Path != path {
			continue
		}
		if rule.Method != "" && rule.Method != method {
			continue
		}

		return rule
	}
	return nil
}

func (cp *ChaosProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rule := cp.findRule(r.URL.Path, r.Method)
	chaosApplied := "none"

	if rule != nil {
		// Apply delay
		if rule.Delay > 0 {
			select {
			case <-time.After(rule.Delay):

			case <-r.Context().Done():
				log.Println("Request cancelled while delaying")
				return
			}

			chaosApplied = "delay"

		}

		// Apply random failure
		if rule.FailureRate > 0 && rand.Float64() < rule.FailureRate {

			if rule.StatusCode == 0 {
				rule.StatusCode = http.StatusInternalServerError
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(rule.StatusCode)
			_, _ = w.Write([]byte(rule.ErrorBody))
			log.Printf("[chaos] injected failed response: %d, body: %s", rule.StatusCode, rule.ErrorBody)

			return
		}
	}
	w.Header().Set("X-Chaos-Applied", chaosApplied)
	cp.Proxy.ServeHTTP(w, r)

}

func (cp *ChaosProxy) Start() error {
	addr := fmt.Sprintf(":%d", cp.Port)

	fmt.Println("Chaos proxy running at", addr)
	return http.ListenAndServe(addr, cp)

}
