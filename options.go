package connect_go_prometheus

type options struct {
	withHistogram bool
}

type InterceptorOption func(opts *options)

func WithHistograms(b bool) InterceptorOption {
	return func(opts *options) {
		opts.withHistogram = b
	}
}

var defaultOptions = &options{
	withHistogram: false,
}

func evaluteOptions(opts ...InterceptorOption) *options {
	for _, opt := range opts {
		opt(defaultOptions)
	}
	return defaultOptions
}
