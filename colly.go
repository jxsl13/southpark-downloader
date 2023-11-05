package main

import (
	"context"

	"github.com/gocolly/colly/v2"
)

func NewCollector(ctx context.Context, userAgent string) *colly.Collector {
	co := colly.NewCollector()
	co.WithTransport(NewContextTransport(ctx))
	co.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
			return
		default:
		}
	})

	co.AllowURLRevisit = false
	co.UserAgent = userAgent

	return co

}
