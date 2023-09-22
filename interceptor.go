package connect_go_prometheus

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
)

const (
	CodeOk = "ok"
)

func NewInterceptor(opts ...InterceptorOption) *Interceptor {
	options := evaluteInterceptorOptions(&interceptorOptions{
		client: DefaultClientMetrics,
		server: DefaultServerMetrics,
	}, opts...)

	return &Interceptor{
		client: options.client,
		server: options.server,
	}
}

var _ connect.Interceptor = (*Interceptor)(nil)

type Interceptor struct {
	client *Metrics
	server *Metrics
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Short-circuit, not configured to report for either client or server.
		if i.client == nil && i.server == nil {
			return next(ctx, req)
		}

		now := time.Now()
		callType := steamTypeString(req.Spec().StreamType)
		callPackage, callMethod := procedureToPackageAndMethod(req.Spec().Procedure)

		var reporter *Metrics
		if req.Spec().IsClient {
			reporter = i.client
		} else {
			reporter = i.server
		}

		if reporter != nil {
			reporter.ReportStarted(callType, callPackage, callMethod)
		}

		resp, err := next(ctx, req)
		code := codeOf(err)

		if reporter != nil {
			reporter.ReportHandled(callType, callPackage, callMethod, code)
			reporter.ReportHandledSeconds(callType, callPackage, callMethod, code, time.Since(now).Seconds())
		}

		return resp, err
	})
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, s)
		if i.client != nil {
			conn = newStreamingClientConn(conn, i)
		}
		return conn
	})
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		var newShc *streamingHandlerConn
		if i.server != nil {
			newShc = newStreamingHandlerConn(shc, i)
			shc = newShc
		}
		err := next(ctx, shc)
		if newShc != nil {
			newShc.reportHandled(err)
		}
		return err
	})
}

func procedureToPackageAndMethod(procedure string) (string, string) {
	procedure = strings.TrimPrefix(procedure, "/") // remove leading slash
	if i := strings.Index(procedure, "/"); i >= 0 {
		return procedure[:i], procedure[i+1:]
	}

	return "unknown", "unknown"
}

func steamTypeString(st connect.StreamType) string {
	switch st {
	case connect.StreamTypeUnary:
		return "unary"
	case connect.StreamTypeClient:
		return "client_stream"
	case connect.StreamTypeServer:
		return "server_stream"
	case connect.StreamTypeBidi:
		return "bidi"
	default:
		return "unknown"
	}
}

func codeOf(err error) string {
	if err == nil {
		return CodeOk
	}
	code := connect.CodeOf(err)
	if code == connect.CodeUnknown {
		if errors.Is(err, context.Canceled) {
			code = connect.CodeCanceled
		} else if errors.Is(err, context.DeadlineExceeded) {
			code = connect.CodeDeadlineExceeded
		}
	}
	return code.String()
}

type interceptorOptions struct {
	client *Metrics
	server *Metrics
}

type InterceptorOption func(*interceptorOptions)

func WithClientMetrics(m *Metrics) InterceptorOption {
	return func(io *interceptorOptions) {
		io.client = m
	}
}

func WithServerMetrics(m *Metrics) InterceptorOption {
	return func(io *interceptorOptions) {
		io.server = m
	}
}

func evaluteInterceptorOptions(defaults *interceptorOptions, opts ...InterceptorOption) *interceptorOptions {
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}
