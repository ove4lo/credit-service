package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	applicationsReceived = promauto.NewCounter(prometheus.CounterOpts{ // creates a counter
		Name: "credit_applications_received_total", // metric name
		Help: "Total number of credit applications received.", // description, displayed in interfaces
	})
)