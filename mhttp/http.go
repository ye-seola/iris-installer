package mhttp

import (
	"context"
	"net"
	"net/http"
	"time"
)

func CreateHTTPClient() *http.Client {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout: time.Millisecond * 10000,
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	dialer := &net.Dialer{
		Timeout:  time.Millisecond * 10000,
		Resolver: r,
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
}
