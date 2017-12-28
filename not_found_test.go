package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundGetShouldReturnNil(t *testing.T) {
	nf := &NotFound{}
	resp, ok := nf.Get("test")
	assert.Equal(t, nil, resp)
	assert.Equal(t, false, ok)
}
