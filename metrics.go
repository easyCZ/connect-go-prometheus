package connect_go_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

func NewServerMetrics(opts ...InterceptorOption) *ServerMetrics {
	config := evaluteOptions(opts...)
	m := &ServerMetrics{
		requestStartedCounter: prom.NewCounterVec(prom.CounterOpts{
			Name: "connect_server_started_total",
			Help: "Total number of RPCs started on the server.",
		}, []string{"type", "service", "method"}),
		requestHandledCounter: prom.NewCounterVec(prom.CounterOpts{
			Name: "connect_server_handled_total",
			Help: "Total number of RPCs started on the server.",
		}, []string{"type", "service", "method", "code"}),
	}

	if config.withHistogram {
		m.serverHandledHistogram = prom.NewHistogramVec(prom.HistogramOpts{
			Name: "connect_server_handled_seconds",
		}, []string{"type", "method", "code"})
	}

	return m
}

type ServerMetrics struct {
	requestStartedCounter  *prom.CounterVec
	requestHandledCounter  *prom.CounterVec
	serverHandledHistogram *prom.HistogramVec
}

// type ClientMetrics struct {
// 	requestStarted
// }
