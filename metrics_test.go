package connect_go_prometheus

import (
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestServerMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	sm := NewServerMetrics(
		WithHistogram(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{0.5, 1, 1.5}),
	)
	require.NoError(t, sm.Register(reg))

	started := sm.requestStarted.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	started.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(started))

	handled := sm.requestHandled.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handled.Inc()
	require.EqualValues(t, 1, testutil.ToFloat64(handled))

	if sm.requestHandledSeconds != nil {
		handledHist := sm.requestHandledSeconds.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
		handledHist.Observe(1)

		err := testutil.CollectAndCompare(sm.requestHandledSeconds, strings.NewReader(`
			# HELP namespace_subsystem_connect_server_handled_seconds Histogram of RPCs handled server-side
			# TYPE namespace_subsystem_connect_server_handled_seconds histogram
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="0.5"} 0
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1"} 1
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1.5"} 1
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="Inf"} 1
			namespace_subsystem_connect_server_handled_seconds_sum{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
			namespace_subsystem_connect_server_handled_seconds_count{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
		`))
		require.NoError(t, err)
	}
}

func TestClientMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	sm := NewClientMetrics(
		WithHistogram(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{0.5, 1, 1.5}),
	)
	require.NoError(t, sm.Register(reg))

	started := sm.requestStarted.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	started.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(started))

	handled := sm.requestHandled.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handled.Inc()
	require.EqualValues(t, 1, testutil.ToFloat64(handled))

	if sm.requestHandledSeconds != nil {
		handledHist := sm.requestHandledSeconds.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
		handledHist.Observe(1)

		err := testutil.CollectAndCompare(sm.requestHandledSeconds, strings.NewReader(`
			# HELP namespace_subsystem_connect_client_handled_seconds Histogram of RPCs handled client-side
			# TYPE namespace_subsystem_connect_client_handled_seconds histogram
			namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="0.5"} 0
			namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1"} 1
			namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1.5"} 1
			namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="Inf"} 1
			namespace_subsystem_connect_client_handled_seconds_sum{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
			namespace_subsystem_connect_client_handled_seconds_count{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
		`))
		require.NoError(t, err)
	}
}