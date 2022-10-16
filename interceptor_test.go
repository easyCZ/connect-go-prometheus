package connect_go_prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	greetv1 "github.com/easyCZ/connect-go-prometheus/gen/greet"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestNewInterceptor_ServerInterceptor_Unary(t *testing.T) {
	interceptor := NewInterceptor()

	_, client := setup(t, interceptor)

	_, err := client.Greet(context.Background(), connect.NewRequest(&greetv1.GreetRequest{Name: "foo"}))
	require.NoError(t, err)

	require.EqualValues(t, 1, testutil.ToFloat64(interceptor.server.requestStartedCounter.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")))
	require.EqualValues(t, 1, testutil.ToFloat64(interceptor.server.requestHandledCounter.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", "ok")))
}

func setup(t *testing.T, interceptor *Interceptor) (*httptest.Server, greetconnect.GreetServiceClient) {
	_, handler := greetconnect.NewGreetServiceHandler(&GreetHandler{}, connect.WithHandlerOptions(connect.WithInterceptors(interceptor)))
	srv := httptest.NewServer(handler)

	t.Cleanup(func() {
		srv.Close()
	})

	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL)

	return srv, client
}

type GreetHandler struct {
	greetconnect.UnimplementedGreetServiceHandler
}

func (h *GreetHandler) Greet(ctx context.Context, req *connect.Request[greetv1.GreetRequest]) (*connect.Response[greetv1.GreetResponse], error) {
	return connect.NewResponse(&greetv1.GreetResponse{Greeting: "hello"}), nil
}
