package badgerraftcache

import (
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/sniperkit/cache/core"
	"github.com/sniperkit/configer"
	"github.com/sniperkit/vipertags"
)

type badgerRaftCacheConfig struct {
	Provider       string        `json:"provider" config:"database.provider"`
	Endpoints      []string      `json:"endpoints" config:"database.endpoints"`
	MaxConnections int           `json:"max_connections" config:"database.max_connections" default:"0"`
	done           chan struct{} `json:"-" config:"-"`
}

// Config ...
var (
	PluginConfig = &badgerRaftCacheConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (badgerRaftCacheConfig) ConfigName() string {
	return "BadgerRaftKV"
}

// SetDefaults ...
func (a *badgerRaftCacheConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *badgerraftcacheConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	if a.Provider == "" {
		a.Provider = a.ConfigName()
	}
	if a.MaxConnections == 0 {
		a.MaxConnections = core.DefaultMaxConnections
	}
}

// Read several config files (yaml, json or env variables)
func (a *badgerRaftCacheConfig) Configor(files []string) {
	configor.Load(&PluginConfig, files...)
}

// Wait ...
func (c badgerRaftCacheConfig) Wait() {
	<-c.done
}

// String ...
func (c badgerRaftCacheConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c badgerRaftCacheConfig) Debug() {
	// log.Debug("BadgerRaft-KV PluginConfig = ", c)
}

func init() {
	config.Register(PluginConfig)
}
