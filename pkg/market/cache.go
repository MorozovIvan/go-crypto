package market

import (
	"sync"
	"time"
)

// CacheEntry represents a cached data entry
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
}

// Cache represents the caching system
type Cache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores data in cache with TTL
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// Get retrieves data from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		return nil, false
	}

	return entry.Data, true
}

// Delete removes an entry from cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	totalEntries := len(c.entries)
	expiredEntries := 0

	for _, entry := range c.entries {
		if time.Since(entry.Timestamp) > entry.TTL {
			expiredEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":   totalEntries,
		"expired_entries": expiredEntries,
		"active_entries":  totalEntries - expiredEntries,
	}
}

// cleanup removes expired entries periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			for key, entry := range c.entries {
				if time.Since(entry.Timestamp) > entry.TTL {
					delete(c.entries, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}

// MarketDataCache represents cached market data
type MarketDataCache struct {
	Value       interface{} `json:"value"`
	Indicator   string      `json:"indicator"`
	Score       float64     `json:"score"`
	ChartData   []float64   `json:"chart_data"`
	ChartLabels []string    `json:"chart_labels"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Global cache instance
var globalCache *Cache

func init() {
	globalCache = NewCache()
}

// GetCachedMarketData retrieves cached market data
func GetCachedMarketData(metric string) (*MarketDataCache, bool) {
	data, exists := globalCache.Get("market_" + metric)
	if !exists {
		return nil, false
	}

	cached, ok := data.(*MarketDataCache)
	if !ok {
		return nil, false
	}

	return cached, true
}

// SetCachedMarketData stores market data in cache
func SetCachedMarketData(metric string, data *MarketDataCache, ttl time.Duration) {
	globalCache.Set("market_"+metric, data, ttl)
}

// GetCacheStats returns cache statistics
func GetCacheStats() map[string]interface{} {
	return globalCache.GetStats()
}

// ClearCache clears all cached data
func ClearCache() {
	globalCache.Clear()
}
