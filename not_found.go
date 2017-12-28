package cache

import (
	"errors"
)

// NopCache provides a no-op cache implementation that doesn't actually cache anything.
// var NotFound = new(notFound)

type NotFound struct{}

func (c *NotFound) Get(string) ([]byte, bool) { return nil, false }
func (c *NotFound) Set(string, []byte)        {}
func (c *NotFound) Delete(string)             {}
func (c *NotFound) Debug(action string)       {}
func (c *NotFound) Action(name string, args ...interface{}) (map[string]*interface{}, error) {
	return nil, errors.New("Action not implemented yet")
}
