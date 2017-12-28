package badgerrediscache

import (
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zaibon/resp"
)

const (
	defaultCacheValueDir string = "httpcache"
	defaultCacheDir      string = "./shared/data/cache/.badgerredis"
)

var defaultCachePath string = filepath.Join(defaultCacheDir, defaultCacheValueDir)

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	sync.RWMutex
	db          *badger.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	StoragePath string
	ValueDir    string
	SyncWrites  bool
	Debug       bool
}

func New(config *Config) (*Cache, error) {
	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("badgerrediscache.New()")
	}
	badgerConfig := badger.DefaultOptions
	if config == nil {
		badgerConfig.Dir = defaultCacheDir
		badgerConfig.ValueDir = defaultCacheValueDir
		badgerConfig.SyncWrites = false
	} else {
		badgerConfig.Dir = config.StoragePath
		badgerConfig.ValueDir = config.ValueDir
		badgerConfig.SyncWrites = config.SyncWrites
	}
	client, err := badger.Open(badgerConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
		}).Fatalln("badgerrediscache.New().badger.Open(), ERROR: ", err)
		return nil, err
	}
	return &Cache{
		db: client,
	}, nil
}

func Mount(client *badger.DB) *Cache {
	return &Cache{db: client}
}

func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.Lock()
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		resp, err := item.Value()
		if err != nil {
			return err
		}
		if c.debug {
			log.WithFields(log.Fields{
				"resp": resp,
			}).Debug("badgerrediscache.Get()")
		}
		return nil
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debug("badgerrediscache.Get()")
	}
	return resp, err != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(resp))
		return err
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debugln("badgerrediscache.Set()")
	}
	return
}

// Delete key from the cache
func (c *Cache) Delete(key string) {
	c.Lock()
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"key": key,
			"ok":  err != nil,
		}).Debug("badgerrediscache.Delete()")
	}
	return
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}
