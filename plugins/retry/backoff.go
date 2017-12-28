package cacher

import (
	"time"

	"github.com/cenkalti/backoff"
)

/*
	Refs:
	- github.com/bluele/greq
	- https://github.com/technoweenie/httpretry
	- https://github.com/jianyuan/go-playground/blob/master/retry/retry.go
	- https://github.com/gravitational/coordinate/blob/master/leader/backoff.go
	- https://github.com/ContainX/go-springcloud/blob/master/discovery/eureka/backoff.go
*/

func newExponentialBackOff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 5 * time.Second
	b.Multiplier = 2.0
	b.MaxInterval = 320 * time.Second
	b.Reset()
	return b
}

func newAggressiveExponentialBackOff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Minute
	b.Multiplier = 2.0
	b.MaxInterval = 16 * time.Minute
	b.Reset()
	return b
}
