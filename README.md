# connect-go-prometheus
[Prometheus](https://prometheus.io/) monitoring for [connect-go](https://github.com/bufbuild/connect-go).

## Interceptors
This library defines [interceptors](https://connect.build/docs/go/interceptors) to observe both client-side and server-side calls.

## Install
```bash
go get -u github.com/easyCZ/connect-go-prometheus
```

## Usage
```golang
import (
    "github.com/easyCZ/connect-go-prometheus"
)

// Construct the interceptor. The same intereceptor is used for both client-side and server-side.
interceptor := connect_go_prometheus.NewInterceptor()

// Use the interceptor when constructing a new connect-go handler
_, _ := your_connect_package.NewServiceHandler(handler, connect.WithInterceptors(interceptor))

// Or with a client
client := your_connect_package.NewServiceClient(http.DefaultClient, serverURL, connect.WithInterceptors(interceptor))
```
For configuration, and more advanced use cases see [Configuration](#Configuration)

## Metrics

Metrics exposed use the following labels:
* `type` - one of `unary`, `client_stream`, `server_stream` or `bidi`
* `service` - name of the service, for example `myservice.greet.v1`
* `method` - name of the method, for example `SayHello`
* `code` - the resulting outcome of the RPC. The codes match [connect-go Error Codes](https://connect.build/docs/protocol#error-codes) with the addition of `ok` for succesful RPCs. 


### Server-side metrics
* Counter `connect_server_started_total` with `(type, service, method)` labels
* Counter `connect_server_handled_total` with `(type, service, method, code)` labels
* (optionally) Histogram `connect_server_handled_seconds` with `(type, service, method, code)` labels

### Client-side metrics
* Counter `connect_client_started_total` with `(type, service, method)` labels
* Counter `connect_client_handled_total` with `(type, service, method, code)` labels
* (optionally) Histogram `connect_client_handled_seconds` with `(type, service, method, code)` labels

## Configuration

### Customizing client/server metrics reported
```golang
import (
    "github.com/easyCZ/connect-go-prometheus"
    prom "github.com/prometheus/client_golang/prometheus"
)

options := []connect_go_prometheus.MetricOption{
    connect_go_prometheus.WithHistogram(true),
    connect_go_prometheus.WithNamespace("namespace"),
    connect_go_prometheus.WithSubsystem("subsystem"),
    connect_go_prometheus.WithConstLabels(prom.Labels{"component": "foo"}),
    connect_go_prometheus.WithHistogramBuckets([]float64{1, 5}),
}

// Construct client metrics
clientMetrics := connect_go_prometheus.NewClientMetrics(options...)

// Construct server metrics
serverMetrics := connect_go_prometheus.NewServerMetrics(options...)

// When you construct either client/server metrics with options, you must also register the metrics with your Prometheus Registry
prom.MustRegister(clientMetrics, serverMetrics)

// Construct the interceptor with our configured metrics
interceptor := connect_go_prometheus.NewInterceptor(
    connect_go_prometheus.WithClientMetrics(clientMetrics),
    connect_go_prometheus.WithServerMetrics(serverMetrics),
)
```

### Registering metrics against a Registry
You may want to register metrics against a [Prometheus Registry](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Registry). You can do this with the following:
```golang
import (
    "github.com/easyCZ/connect-go-prometheus"
    prom "github.com/prometheus/client_golang/prometheus"
)

clientMetrics := connect_go_prometheus.NewClientMetrics()
serverMetrics := connect_go_prometheus.NewServerMetrics()

registry := prom.NewRegistry()
registry.MustRegister(clientMetrics, serverMetrics)

interceptor := connect_go_prometheus.NewInterceptor(
    connect_go_prometheus.WithClientMetrics(clientMetrics),
    connect_go_prometheus.WithServerMetrics(serverMetrics),
)
```

### Disabling client/server metrics reporting
To disable reporting of either client or server metrics, pass `nil` as an option.
```golang
import (
    "github.com/easyCZ/connect-go-prometheus"
)

// Disable client-side metrics
interceptor := connect_go_prometheus.NewInterceptor(
    connect_go_prometheus.WithClientMetrics(nil),
)

// Disable server-side metrics
interceptor := connect_go_prometheus.NewInterceptor(
    connect_go_prometheus.WithServerMetrics(nil),
)
```

test
