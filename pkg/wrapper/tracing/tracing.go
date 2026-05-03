package tracing

type Provider struct{}

func NewProvider(service string) *Provider { return &Provider{} }
