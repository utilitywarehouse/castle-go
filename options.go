package castle

type options struct {
	metricsEnabled bool
}

type Opt func(*options)

func WithMetrics(b bool) Opt {
	return func(o *options) {
		o.metricsEnabled = b
	}
}
