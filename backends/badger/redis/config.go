package badgerrediscache

import (
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/sniperkit/cache/core"
	"github.com/sniperkit/configer"
	"github.com/sniperkit/vipertags"
)

type badgerRedisCacheConfig struct {
	Provider       string        `json:"provider" config:"database.provider"`
	Endpoints      []string      `json:"endpoints" config:"database.endpoints"`
	MaxConnections int           `json:"max_connections" config:"database.max_connections" default:"0"`
	done           chan struct{} `json:"-" config:"-"`
}

// Config ...
var (
	PluginConfig = &badgerRedisCacheConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (badgerRedisCacheConfig) ConfigName() string {
	return "BadgerRedisKV"
}

// SetDefaults ...
func (a *badgerRedisCacheConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *badgerRedisCacheConfig) Read() {
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
func (a *badgerRedisCacheConfig) Configor(files []string) {
	configor.Load(&PluginConfig, files...)
}

// Wait ...
func (c badgerRedisCacheConfig) Wait() {
	<-c.done
}

// String ...
func (c badgerRedisCacheConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c badgerRedisCacheConfig) Debug() {
	// log.Debug("BadgerRedisKV PluginConfig = ", c)
}

func init() {
	config.Register(badgerRedisCacheConfig)
}
