package proxy

import "time"

type ChaosRule struct {
	Path        string
	Method      string
	Delay       time.Duration
	FailureRate float64
	StatusCode  int
	ErrorBody   string
}
