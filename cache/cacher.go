package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
)

// ICacher is the interface for cache service
type ICacher interface {
	Set(key string, value interface{}, expire time.Duration) error
	MSet(kv map[string]interface{}) error
	Get(key string) (string, error)
	MGet(keys []string) ([]interface{}, error)
	Expire(key string, expire time.Duration) error
	Expires(keys []string, expire time.Duration) error
	Del(keys ...string) error
	Exists(key string) (bool, error)
	Close() error
}

// ICacherConfig is cacher configuration interface
type ICacherConfig interface {
	Endpoint() string
	Password() string
	DB() int
	ConnectionSettings() ICacherConnectionSettings
}

// ICacherConnectionSettings is connection settings for cacher
type ICacherConnectionSettings interface {
	PoolSize() int
	MinIdleConns() int
	MaxRetries() int
	MinRetryBackoff() time.Duration
	MaxRetryBackoff() time.Duration
	IdleTimeout() time.Duration
	IdleCheckFrequency() time.Duration
	PoolTimeout() time.Duration
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
}

// DefaultCacherConnectionSettings contains default connection settings, this intend to use as embed struct
type DefaultCacherConnectionSettings struct{}

func NewDefaultCacherConnectionSettings() ICacherConnectionSettings {
	return &DefaultCacherConnectionSettings{}
}

func (setting *DefaultCacherConnectionSettings) PoolSize() int {
	return 50
}

func (setting *DefaultCacherConnectionSettings) MinIdleConns() int {
	return 5
}

func (setting *DefaultCacherConnectionSettings) MaxRetries() int {
	return 3
}

func (setting *DefaultCacherConnectionSettings) MinRetryBackoff() time.Duration {
	return 10 * time.Millisecond
}

func (setting *DefaultCacherConnectionSettings) MaxRetryBackoff() time.Duration {
	return 500 * time.Millisecond
}

func (setting *DefaultCacherConnectionSettings) IdleTimeout() time.Duration {
	return 30 * time.Minute
}

func (setting *DefaultCacherConnectionSettings) IdleCheckFrequency() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) PoolTimeout() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) ReadTimeout() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) WriteTimeout() time.Duration {
	return time.Minute
}

// Cacher is the struct for cache service
type Cacher struct {
	config      ICacherConfig
	clientMutex sync.Mutex
	client      *redis.Client
	oldClients  []*redis.Client
	subsribers  *sync.Map
}

// NewCacher return new Cacher
func NewCacher(config ICacherConfig) *Cacher {
	return &Cacher{
		config:     config,
		oldClients: nil,
		subsribers: &sync.Map{},
	}
}

func (cache *Cacher) newClient() *redis.Client {
	cfg := cache.config
	settings := cfg.ConnectionSettings()
	return redis.NewClient(&redis.Options{
		Addr:               cfg.Endpoint(),
		Password:           cfg.Password(),
		DB:                 cfg.DB(),
		PoolSize:           settings.PoolSize(),
		MinIdleConns:       settings.MinIdleConns(),
		MaxRetries:         settings.MaxRetries(),
		MinRetryBackoff:    settings.MinRetryBackoff(),
		MaxRetryBackoff:    settings.MaxRetryBackoff(),
		IdleTimeout:        settings.IdleTimeout(),
		IdleCheckFrequency: settings.IdleCheckFrequency(),
		PoolTimeout:        settings.PoolTimeout(),
		ReadTimeout:        settings.ReadTimeout(),
		WriteTimeout:       settings.WriteTimeout(),
	})
}

func (cache *Cacher) getClient() (*redis.Client, error) {
	cache.clientMutex.Lock()
	defer cache.clientMutex.Unlock()

	retriesDelayMs := cache.getRetriesDelayInMs()
	retries := -1
	for {
		retries++
		if retries > len(retriesDelayMs)-1 {
			return nil, fmt.Errorf("cacher: retry exceed limits")
		}

		client := cache.client
		if client == nil {
			client = cache.newClient()
			cache.client = client
		}

		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			// Wait by retry delay then reset client and try connect again
			time.Sleep(time.Millisecond * time.Duration(retriesDelayMs[retries]))
			cache.client = nil
			continue
		}

		// If we can PING without error, just return
		return client, nil
	}
}

// Close close the redis client
func (cache *Cacher) Close() error {
	cache.clientMutex.Lock()
	defer cache.clientMutex.Unlock()

	// Close current client
	client := cache.client
	if client != nil {
		cache.client = nil

		err := client.Close()
		if err != nil {
			return err
		}

		// Close old clients
		for _, client := range cache.oldClients {
			err := client.Close()
			if err != nil {
				return err
			}
		}
		if len(cache.oldClients) > 0 {
			cache.oldClients = nil
		}
	}

	return nil
}

// getRetriesDelayInMs sum only 1 second
func (cache *Cacher) getRetriesDelayInMs() []int {
	return []int{200, 200, 200, 200, 200}
}

// Exists check if key is exists
func (cache *Cacher) Exists(key string) (bool, error) {

	c, err := cache.getClient()
	if err != nil {
		return false, err
	}

	val, err := c.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	// val == 1 means key is exists
	return val == 1, nil
}

// Del the cache by keys
func (cache *Cacher) Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	c, err := cache.getClient()
	if err != nil {
		return err
	}

	// Delete 10000 items per page
	pageLimit := 10000
	from := 0
	to := pageLimit

	for {
		// Lower bound
		if from >= len(keys) {
			break
		}
		// Upper bound
		if to > len(keys) {
			to = len(keys)
		}

		delKeys := keys[from:to]
		if len(delKeys) == 0 {
			break
		}

		_, err = c.Del(context.Background(), delKeys...).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			} else {
				return err
			}
		}
		from += pageLimit
		to += pageLimit
	}

	return nil
}

// Expires set expiration for objects in cache
// if there is error happen, just return last error
func (cache *Cacher) Expires(keys []string, expire time.Duration) error {
	return cache.expires(keys, expire)
}

// Expire set expiration for object in cache
func (cache *Cacher) Expire(key string, expire time.Duration) error {
	return cache.expires([]string{key}, expire)
}

// Expires set expiration for objects in cache
// if there is error happen, just return last error
func (cache *Cacher) expires(keys []string, expire time.Duration) error {
	c, err := cache.getClient()
	if err != nil {
		return err
	}

	var lastErr error
	for _, key := range keys {
		err = c.Expire(context.Background(), key, expire).Err()
		if err != nil {
			if err == redis.Nil {
				// Key does not exists
				return nil
			} else {
				lastErr = err
			}
		}
	}
	return lastErr
}

// MGet get by multiple keys, the value can be nil, so it will return []interface{} instead of []string
func (cache *Cacher) MGet(keys []string) ([]interface{}, error) {

	c, err := cache.getClient()
	if err != nil {
		return nil, err
	}

	vals, err := c.MGet(context.Background(), keys...).Result()
	if err == redis.Nil {
		// Key does not exists
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return vals, nil
}

// Get object from cache
func (cache *Cacher) Get(key string) (string, error) {

	c, err := cache.getClient()
	if err != nil {
		return "", err
	}

	val, err := c.Get(context.Background(), key).Result()
	if err == redis.Nil {
		// Key does not exists
		return "", nil
	} else if err != nil {
		return "", err
	}

	return val, nil
}

// MSet set multiple key value
func (cache *Cacher) MSet(kv map[string]interface{}) error {

	c, err := cache.getClient()
	if err != nil {
		return err
	}

	pairs := []interface{}{}
	for k, v := range kv {

		str, ok := v.(string)
		// Check empty string if value string
		if ok && len(str) == 0 {
			pairs = append(pairs, k, "")
			continue
		}
		// If value is string, not pass it to json.Marshal
		if len(str) > 0 {
			pairs = append(pairs, k, str)
			continue
		}

		strb, err := json.Marshal(v)
		if err != nil {
			return err
		}
		pairs = append(pairs, k, strb)
	}

	err = c.MSet(context.Background(), pairs...).Err()
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cacher) Set(key string, value interface{}, expire time.Duration) error {

	c, err := cache.getClient()
	if err != nil {
		return err
	}

	str, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = c.Set(context.Background(), key, str, expire).Err()
	if err != nil {
		if err == redis.Nil {
			// Key does not exists
			return nil
		} else {
			return err
		}
	}

	return nil
}
