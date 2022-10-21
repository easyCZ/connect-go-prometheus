package connect_go_prometheus

import (
	"context"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
)

var ()

func NewInterceptor(client *Metrics, server *Metrics) *Interceptor {
	return &Interceptor{
		client: client,
		server: server,
	}
}

var _ connect.Interceptor = (*Interceptor)(nil)

type Interceptor struct {
	client *Metrics
	server *Metrics
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
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
			i.client.ReportHandled(callType, callPackage, callMethod, code)
			i.client.ReportHandledSeconds(callType, callPackage, callMethod, code, time.Since(now).Seconds())
		}

		return resp, err
	})
}

func (i *Interceptor) WrapStreamingClient(connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		return nil
	})
}

func (i *Interceptor) WrapStreamingHandler(connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		return nil
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
		return "ok"
	}
	return connect.CodeOf(err).String()
}
