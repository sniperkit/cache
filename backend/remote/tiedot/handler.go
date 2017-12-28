package tiedotcache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/HouzuoGuo/tiedot/dberr"
	log "github.com/sirupsen/logrus"
)

/*
	Refs:
	- https://github.com/HouzuoGuo/tiedot/blob/master/examples/example.go
*/

const (
	defaultStorageBasename   string = "tiedot"
	defaultStoragePrefixPath string = "./shared/data/cache/.tiedot"
	defaultCollectionName    string = "httpcache"
)

var (
	defaultStoragePath string = filepath.Join(defaultStoragePrefixPath, defaultStorageFileExtension)
)

// Cache is an implementation of httpcache.Cache that uses a bolt database.
type Cache struct {
	sync.RWMutex

	db             *db.DB
	collection     *db.Col
	storagePath    string
	collectionName string
	debug          bool
	stats          bool
	compress       bool
}

type Config struct {
	IndexName      string
	CollectionName string
	StoragePath    string
	ReadOnly       bool
	StrictMode     bool
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
		}).Warnf("tiedotcache.New()")
	}

	if config == nil {
		log.Warnln("tiedotcache.New() ERROR: config is nil")
		config.StoragePath = defaultStoragePath
		config.CollectionName = defaultCollectionName
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStoragePath
	}

	if config.CollectionName == "" {
		config.CollectionName = defaultCollectionName
	}

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("tiedotcache.New() ---> post-processed")
	}

	var err error
	cache := &Cache{}
	cache.storagePath = config.StoragePath
	cache.collectionName = config.CollectionName
	cache.compress = config.Compress
	cache.debug = config.Debug
	cache.stats = config.Stats

	// (Create if not exist) open a database
	cache.db, err = db.OpenDB(config.StoragePath)
	if err != nil {
		return nil, err
	}

	// Create item collection
	if err := db.Create(config.CollectionName); err != nil {
		return nil, err
	}
	// Start using a collection (the reference is valid until DB schema changes or Scrub is carried out)
	cache.collection = cache.db.Use(config.CollectionName)

	return cache, nil
}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *db.DB) *Cache {
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	if c.debug {
		log.WithFields(log.Fields{
			"collectionName": c.collectionName,
		}).Warnf("tiedotcache.Close()")
	}
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.RLock()
	data, err := c.collection.Read(key)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
		}).Fatalln("tiedotcache.Get().db.Read(), ERROR: ", err)
	}
	c.RUnlock()
	// data to bytes
	// optional: decompress
	if c.compress {
		var err error
		resp, err = ungzipData(resp)
		if err != nil {
			log.WithFields(log.Fields{
				"config": config,
			}).Fatalln("tiedotcache.Get().ungzipData(), ERROR: ", err)
			return resp, false
		}
	}
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	// Insert document (afterwards the docID uniquely identifies the document and will never change)
	if c.compress {
		var err error
		resp, err = gzipData(resp)
		if err != nil {
			log.WithFields(log.Fields{
				"config": config,
			}).Fatalln("tiedotcache.Set().gzipData(), ERROR: ", err)
			return resp, false
		}
	}
	docID, err := c.collection.Insert(map[string]interface{}{"data": string(resp)})
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
		}).Fatalln("tiedotcache.Set().db.Insert(), ERROR: ", err)
	}
	if c.debug {
		log.WithFields(log.Fields{
			"collectionName": c.collectionName,
			"docID":          docID,
		}).Info("tiedotcache.Delete() OK")
	}
	c.Unlock()
}

func (c *Cache) Debug(action string) {}

func (c *Cache) Action(name string, args ...interface{}) (map[string]*interface{}, error) {
	return nil, errors.New("Action not implemented yet")
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	c.Lock()
	// More complicated error handing - identify the error Type. check the error code if the the document no longer exists.
	if err := c.collection.Delete(key); dberr.Type(err) == dberr.ErrorNoDoc {
		if c.debug {
			log.WithFields(log.Fields{
				"collectionName": c.collectionName,
				"key":            key,
			}).Warnln("tiedotcache.Set().db.Delete(), ERROR: ", err)
		}
	}
	c.Unlock()
	if c.debug {
		log.WithFields(log.Fields{
			"collectionName": c.collectionName,
			"key":            key,
		}).Info("tiedotcache.Delete() OK")
	}
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
