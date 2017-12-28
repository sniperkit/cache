package cache

import (
	"net/http"
	"strconv"

	"github.com/sniperkit/logger"
)

type CustomTransport struct {
	Transport http.RoundTripper
	MaxStale  int //seconds
	Debug     bool
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	req2 := cloneRequest(req) // per RoundTripper contract
	req2.Header.Add("Cache-Control", "Max-Stale="+strconv.Itoa(c.MaxStale))

	logx.DebugWithFields(logger.Fields{
		"Max-Stale": c.MaxStale,
	}, "httpcache.CustomTransport.RoundTrip()")

	res, err := transport.RoundTrip(req2)
	return res, err
}
