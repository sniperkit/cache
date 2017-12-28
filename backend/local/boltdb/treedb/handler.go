package boltdbtreecache

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/cellstate/treedb"
	log "github.com/sirupsen/logrus"
)

const (
	bktName string = "httpcache"
	bktDir  string = "./shared/data/cache/.treedb"
)

var defaultCacheDir string = filepath.Join(bktDir, bktName)

type Cache struct {
	sync.RWMutex
	db          *bolt.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName  string
	StoragePath string
	Debug       bool
}

func New(config *Config) (*Cache, error) {

	if config == nil {
		log.Println("boltdbtreecache.New(): config is nil")
		config.StoragePath = defaultCacheDir
	}

	if config.StoragePath == "" {
		return nil, errors.New("boltdbtreecache.New(): Storage path is not defined.")
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	cache := &Cache{}
	cache.debug = config.Debug

	var err error
	cache.db, err = bolt.Open(config.StoragePath, 0600, nil)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbtreecache.New(): Open error: %v", err)
		}
		return nil, err
	}

	init := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(config.BucketName))
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbtreecache.New(): CreateBucketIfNotExists error: %v", err)
		}
		return err
	}

	if err := cache.db.Update(init); err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbtreecache.New(): Update error: %v", err)
		}
		if err := cache.db.Close(); err != nil {
			if config.Debug {
				log.WithFields(log.Fields{
					"config": config,
					"cache":  cache,
				}).Fatalf("boltdbtreecache.New(): Close error: %v", err)
			}
		}
		return nil, err
	}

	return cache, nil

}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *bolt.DB) *Cache {
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	// c.RLock()
	// defer c.RUnlock()

	get := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		resp = bkt.Get([]byte(key))
		return nil
	}
	if err := c.db.View(get); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"get":        get != nil,
			}).Fatalf("boltdbtreecache.Get() View ERROR: ", err)
		}
		return resp, false
	}
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"get":        get != nil,
			"ok":         resp != nil,
		}).Info("boltdbtreecache.Get() OK")
	}
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	// c.RLock()
	// defer c.RUnlock()
	// strconv.FormatUint(u.ID, 10)

	set := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		return bkt.Put([]byte(key), resp)
	}
	if err := c.db.Update(set); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"set":        set != nil,
			}).Fatalf("boltdbtreecache.Set() Update ERROR: ", err)
		}
	}
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         set != nil,
		}).Info("boltdbtreecache.Set() OK")
	}
}

func (c *Cache) Debug(action string) {

}

func (c *Cache) Action(name string, args ...interface{}) (map[string]*interface{}, error) {
	return nil, errors.New("Action not implemented yet")
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	// c.RLock()
	// defer c.RUnlock()

	del := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		return bkt.Delete([]byte(key))
	}
	if err := c.db.Update(del); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"ok":         del != nil,
			}).Fatalf("boltdbtreecache.Delete() Update ERROR: ", err)
		}
	}
	if c.debug {
		log.WithFields(log.Fields{
			"context":    "Entry",
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         del != nil,
		}).Info("boltdbtreecache.Set() OK")
	}
}

// Ping connects to the database. Returns nil if successful.
func (c *Cache) Ping() error {
	return c.db.View(func(tx *bolt.Tx) error { return nil })
}
