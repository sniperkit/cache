//
package bboltdbcache

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"bytes"
	"compress/gzip"
	"io/ioutil"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	// "github.com/sniperkit/cache/helpers"
)

/*
	Refs:
	- https://github.com/br0xen/boltbrowser
	- https://github.com/aerth/fforum/blob/master/forum.go
	- https://github.com/Everlag/poeitemstore/blob/master/stash/stash.go (encoding json)
	- https://github.com/joltdb/jolt/blob/master/api.go
	- https://github.com/gqf2008/btwan/blob/master/cmd/migrate/migrate.go
*/

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.bbolt"
	defaultStorageFileExtension string = ".bbolt"
)

var (
	defaultStorageDir  string = filepath.Join(bktDir, bktName)
	defaultStorageFile string = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

// Cache is an implementation of httpcache.Cache that uses a bolt database.
type Cache struct {
	// sync.Mutex
	sync.RWMutex
	db          *bolt.DB
	storagePath string
	bucketName  string
	debug       bool
	stats       bool
	compress    bool
	// bucket      *bolt.Bucket
}

type Config struct {
	BucketName     string
	StoragePath    string
	ReadOnly       bool
	StrictMode     bool
	NoSync         bool
	NoFreelistSync bool
	NoGrowSync     bool
	MaxBatchSize   bool
	MaxBatchDelay  bool
	AllocSize      bool
	Compress       bool
	Debug          bool
	Stats          bool
}

type Check struct {
	Enabled   bool
	Key       string
	Requests  int
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiredAt time.Time
	Priority  bool
	Provider  string
}

// https://github.com/gqf2008/btwan/blob/master/cmd/migrate/migrate.go

// New returns a new Cache that uses a bolt database at the given path.
func New(config *Config) (*Cache, error) {

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("bboltcache.New()")
	}

	if config == nil {
		log.Warnln("config is nil")
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStorageFile
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("bboltcache.New() ---> post-processed")
	}

	var err error
	cache := &Cache{}
	cache.storagePath = config.StoragePath
	cache.bucketName = config.BucketName
	cache.compress = config.Compress
	cache.debug = config.Debug
	cache.stats = config.Stats

	cache.db, err = bolt.Open(config.StoragePath, 0755, nil)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("bboltcache.New(): Open error: %v", err)
		}
		return nil, err
	}

	init := func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(config.BucketName))
		if cache.stats {
			log.Printf("Connected to DB(%s) : %+v", config.BucketName, bucket.Stats())
		}
		return err
	}

	if err := cache.db.Update(init); err != nil {
		if config.Debug {
			log.Fatalf("bboltcache.New(): init error: %v", err)
		}
		if err := cache.db.Close(); err != nil {
			if config.Debug {
				log.Fatalf("bboltcache.New(): close error: %v", err)
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
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
		}).Warnf("bboltcache.Close()")
	}
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.RLock()
	get := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		if c.stats {
			log.Printf("Connected to DB(%s) : %+v", c.bucketName, bkt.Stats())
		}
		resp = bkt.Get([]byte(key))
		return nil
	}
	if err := c.db.View(get); err != nil {
		log.Printf("boltdbcache.Get(): view error: %v", err)
		return resp, false
	}
	c.RUnlock()
	if c.compress {
		var err error
		resp, err = ungzipData(resp)
		if err != nil {
			return resp, false
		}
	}
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	set := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		if c.stats {
			log.Printf("Connected to DB(%s) : %+v", c.bucketName, bkt.Stats())
		}
		if c.compress {
			var err error
			resp, err = gzipData(resp)
			if err != nil {
				return errors.New("error while compressing content...")
			}
		}
		return bkt.Put([]byte(key), resp)
	}
	c.Unlock()
	if err := c.db.Update(set); err != nil {
		log.Printf("boltdbcache.Set(): update error: %v", err)
	}
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	c.Lock()
	del := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			if c.debug {
				log.WithFields(log.Fields{
					"context":    "Bucket",
					"bucketName": c.bucketName,
					"key":        key,
				}).Fatal("bboltcache.Delete() Bucket error")
			}
			return errors.New(fmt.Sprintf("bboltcache.Delete(): could not reach the bucket: %s", c.bucketName))
		}
		if c.stats {
			log.Printf("Connected to DB(%s) : %+v", c.bucketName, bkt.Stats())
		}
		return bkt.Delete([]byte(key))
	}
	if err := c.db.Update(del); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"context":    "Delete",
				"bucketName": c.bucketName,
				"key":        key,
			}).Fatalln("bboltcache.Delete() Update: ", err)
		}
		return
	}
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"context":    "Update",
			"bucketName": c.bucketName,
			"key":        key,
		}).Info("bboltcache.Delete() OK")
	}
}

func (c *Cache) Debug(action string) {}

func (c *Cache) Action(name string, args ...interface{}) (map[string]*interface{}, error) {
	return nil, errors.New("Action not implemented yet")
}

func ungzipData(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func gzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func fmtToJsonArr(s []byte) []byte {
	s = bytes.Replace(s, []byte("{"), []byte("[{"), 1)
	s = bytes.Replace(s, []byte("}"), []byte("},"), -1)
	s = bytes.TrimSuffix(s, []byte(","))
	s = append(s, []byte("]")...)
	return s
}
