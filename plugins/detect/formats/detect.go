package formats

import (
	"io"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"
)

const DefaultPeekSize = DefaultBlockSize * 2

type Detector struct {
	mu    sync.RWMutex
	debug bool
}

func (d *Detector) Register(debug *bool) {
	d.debug = &debug
	log.WithFields(log.Fields{
		"d.debug": d.debug,
	}).Info("multigz.Register()")
}

func (d *Detector) Detect(r io.ReadSeeker, peeksize int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return IsProbablyMultiGzip(r, peeksize)
}

func (d *Detector) Debug(action string) {
	c.mu.Lock()
	defer c.mu.Unlock()
}

func (d *Detector) Debug(action string) {
	c.mu.Lock()
	defer c.mu.Unlock()
}
