package cacher

// A Cache interface is used by the Transport to store and retrieve responses.
type Cache interface {
	Get(key string) (responseBytes []byte, ok bool) // Get returns the []byte representation of a cached response and a bool set to true if the value isn't empty
	Set(key string, responseBytes []byte)           // Set stores the []byte representation of a response against a key
	Delete(key string)                              // Delete removes the value associated with the key
	Debug(action string)                            // Delete removes the value associated with the key
	Action(name string, args ...interface{}) (map[string]*interface{}, error)
}

type cacheControl map[string]string
