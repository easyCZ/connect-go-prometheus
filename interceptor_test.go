package connect_go_prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/easyCZ/connect-go-prometheus/gen/greet"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestInterceptor_WithHistogram(t *testing.T) {
	reg := prom.NewRegistry()
	opts := []MetricsOption{
		WithHistogram(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{1, 5}),
	}
	clientMetrics := NewClientMetrics(opts...)
	serverMetrics := NewServerMetrics(opts...)

	reg.MustRegister(clientMetrics, serverMetrics)

	intereceptor := NewInterceptor(clientMetrics, serverMetrics)

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(intereceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(intereceptor))
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

func TestInterceptor_WithoutHistogram(t *testing.T) {
	reg := prom.NewRegistry()
	opts := []MetricsOption{
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
	}
	clientMetrics := NewClientMetrics(opts...)
	serverMetrics := NewServerMetrics(opts...)

	reg.MustRegister(clientMetrics, serverMetrics)

	intereceptor := NewInterceptor(clientMetrics, serverMetrics)

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(intereceptor))
	srv := httptest.NewServer(handler)

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(intereceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "elza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)

	expectedMetrics := []string{
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",

		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(reg, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)
}
