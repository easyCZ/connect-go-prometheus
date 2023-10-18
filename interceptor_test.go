package connect_go_prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/easyCZ/connect-go-prometheus/gen/greet"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

var (
	testMetricOptions = []MetricsOption{
		WithHistogram(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{1, 5}),
	}
)

func TestInterceptor_WithClient_WithServer_Histogram(t *testing.T) {
	reg := prom.NewRegistry()

	clientMetrics := NewClientMetrics(testMetricOptions...)
	serverMetrics := NewServerMetrics(testMetricOptions...)

	reg.MustRegister(clientMetrics, serverMetrics)

	interceptor := NewInterceptor(WithClientMetrics(clientMetrics), WithServerMetrics(serverMetrics))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "elza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)

	expectedMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(reg, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)
}

func TestInterceptor_Default(t *testing.T) {
	interceptor := NewInterceptor()

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "elza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)

	expectedMetrics := []string{
		"connect_client_handled_total",
		"connect_client_started_total",

		"connect_server_handled_total",
		"connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(prom.DefaultGatherer, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)
}

func TestInterceptor_WithClientMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	clientMetrics := NewClientMetrics(testMetricOptions...)
	require.NoError(t, reg.Register(clientMetrics))

	interceptor := NewInterceptor(WithClientMetrics(clientMetrics), WithServerMetrics(nil))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "elza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)

	possibleMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 3, count, "must report only 3 metrics, as server side is disabled")
}

func TestInterceptor_WithServerMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	serverMetrics := NewServerMetrics(testMetricOptions...)
	require.NoError(t, reg.Register(serverMetrics))

	interceptor := NewInterceptor(WithServerMetrics(serverMetrics), WithClientMetrics(nil))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "elza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)

	possibleMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 3, count, "must report only server side metrics, client-side is disabled")
}
