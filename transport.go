package main

import (
	"context"
	"net/http"
)

func NewContextTransport(ctx context.Context) *ContextTransport {
	return &ContextTransport{
		ctx:   ctx,
		trans: http.DefaultTransport,
	}
}

// contextTransport wrapper a context.Context for cancel requests
type ContextTransport struct {
	ctx   context.Context
	trans http.RoundTripper
}

func (t *ContextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(t.ctx)
	return t.trans.RoundTrip(req)
}
