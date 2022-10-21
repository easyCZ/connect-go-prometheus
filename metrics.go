package connect_go_prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// NewServerMetrics creates new Connect metrics for server-side handling.
func NewServerMetrics(opts ...MetricsOption) *Metrics {
	config := evaluateMetricsOptions(&metricsOptions{
		histogramBuckets:          prom.DefBuckets,
		requestStartedName:        "connect_server_started_total",
		requestHandledName:        "connect_server_handled_total",
		requestHandledSecondsName: "connect_server_handled_seconds",
	}, opts...)

	m := &Metrics{
		requestStarted: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestStartedName,
			Help:        "Total number of RPCs started handling server-side",
		}, []string{"type", "service", "method"}),
		requestHandled: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledName,
			Help:        "Total number of RPCs handled server-side",
		}, []string{"type", "service", "method", "code"}),
	}

	if config.withHistogram {
		m.requestHandledSeconds = prom.NewHistogramVec(prom.HistogramOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledSecondsName,
			Help:        "Histogram of RPCs handled server-side",
			Buckets:     config.histogramBuckets,
		}, []string{"type", "service", "method", "code"})
	}

	return m
}

func NewClientMetrics(opts ...MetricsOption) *Metrics {
	config := evaluateMetricsOptions(&metricsOptions{
		histogramBuckets:          prom.DefBuckets,
		requestStartedName:        "connect_client_started_total",
		requestHandledName:        "connect_client_handled_total",
		requestHandledSecondsName: "connect_client_handled_seconds",
	}, opts...)

	m := &Metrics{
		requestStarted: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestStartedName,
			Help:        "Total number of RPCs started handling client-side",
		}, []string{"type", "service", "method"}),
		requestHandled: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledName,
			Help:        "Total number of RPCs handled client-side",
		}, []string{"type", "service", "method", "code"}),
	}

	if config.withHistogram {
		m.requestHandledSeconds = prom.NewHistogramVec(prom.HistogramOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledSecondsName,
			Help:        "Histogram of RPCs handled client-side",
			Buckets:     config.histogramBuckets,
		}, []string{"type", "service", "method", "code"})
	}

	return m
}

type Metrics struct {
	requestStarted        *prom.CounterVec
	requestHandled        *prom.CounterVec
	requestHandledSeconds *prom.HistogramVec
}

// Register registers the ClientMetrics with a prometheus.Registry
func (m *Metrics) Register(registry *prom.Registry) error {
	if err := registry.Register(m.requestStarted); err != nil {
		return err
	}

	if err := registry.Register(m.requestHandled); err != nil {
		return err
	}

	if m.requestHandledSeconds != nil {
		if err := registry.Register(m.requestHandledSeconds); err != nil {
			return err
		}
	}

	return nil
}

func (m *Metrics) ReportStarted(callType, service, method string) {
	m.requestStarted.WithLabelValues(callType, service, method).Inc()
}

func (m *Metrics) ReportHandled(callType, service, method, code string) {
	m.requestHandled.WithLabelValues(callType, service, method, code).Inc()
}

func (m *Metrics) ReportHandledSeconds(callType, service, method, code string, val float64) {
	if m.requestHandledSeconds != nil {
		m.requestHandledSeconds.WithLabelValues(callType, service, method, code).Observe(val)
	}
}

type metricsOptions struct {
	withHistogram    bool
	histogramBuckets []float64

	namespace string
	subsystem string

	requestStartedName        string
	requestHandledName        string
	requestHandledSecondsName string

	constLabels prom.Labels
}

type MetricsOption func(opts *metricsOptions)

func WithHistogram(enabled bool) MetricsOption {
	return func(opts *metricsOptions) {
		opts.withHistogram = enabled
	}
}

func WithHistogramBuckets(buckets []float64) MetricsOption {
	return func(opts *metricsOptions) {
		opts.histogramBuckets = buckets
	}
}

func WithNamespace(namespace string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.namespace = namespace
	}
}

func WithSubsystem(subsystem string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.subsystem = subsystem
	}
}

func withRequestStartedName(name string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.requestStartedName = name
	}
}

func withrequestHandledName(name string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.requestHandledName = name
	}
}

func withRequestedHandledSecondsName(name string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.requestHandledSecondsName = name
	}
}

func WithConstLabels(labels prom.Labels) MetricsOption {
	return func(opts *metricsOptions) {
		opts.constLabels = labels
	}
}

func evaluateMetricsOptions(defaults *metricsOptions, opts ...MetricsOption) *metricsOptions {
	for _, opt := range opts {
		opt(defaults)
	}

	return defaults
}
