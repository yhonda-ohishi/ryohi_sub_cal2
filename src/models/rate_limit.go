package models

import (
	"fmt"
	"sync"
	"time"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	Rate      int      `json:"rate" yaml:"rate"`
	Period    string   `json:"period" yaml:"period"`
	BurstSize int      `json:"burst_size" yaml:"burst_size"`
	KeyType   string   `json:"key_type" yaml:"key_type"`
	WhiteList []string `json:"white_list" yaml:"white_list"`
}

// Validate validates the rate limit configuration
func (r *RateLimitConfig) Validate() error {
	if !r.Enabled {
		return nil
	}
	
	if r.Rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}
	
	validPeriods := []string{"second", "minute", "hour"}
	valid := false
	for _, p := range validPeriods {
		if r.Period == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid period: %s (must be second, minute, or hour)", r.Period)
	}
	
	if r.BurstSize < 0 {
		return fmt.Errorf("burst size cannot be negative")
	}
	
	if r.BurstSize == 0 {
		r.BurstSize = r.Rate // Default burst size equals rate
	}
	
	validKeyTypes := []string{"IP", "API_KEY", "USER_ID", "GLOBAL"}
	valid = false
	for _, kt := range validKeyTypes {
		if r.KeyType == kt {
			valid = true
			break
		}
	}
	if !valid {
		if r.KeyType == "" {
			r.KeyType = "IP" // Default key type
		} else {
			return fmt.Errorf("invalid key type: %s", r.KeyType)
		}
	}
	
	return nil
}

// GetPeriodDuration returns the period as a time.Duration
func (r *RateLimitConfig) GetPeriodDuration() time.Duration {
	switch r.Period {
	case "second":
		return time.Second
	case "minute":
		return time.Minute
	case "hour":
		return time.Hour
	default:
		return time.Minute
	}
}

// IsWhitelisted checks if the given key is whitelisted
func (r *RateLimitConfig) IsWhitelisted(key string) bool {
	for _, wl := range r.WhiteList {
		if wl == key {
			return true
		}
	}
	return false
}

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	config    *RateLimitConfig
	buckets   map[string]*TokenBucket
	mutex     sync.RWMutex
	cleanupAt time.Time
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens    float64
	capacity  float64
	rate      float64
	lastFill  time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:    config,
		buckets:   make(map[string]*TokenBucket),
		cleanupAt: time.Now().Add(1 * time.Hour),
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	if !rl.config.Enabled {
		return true
	}
	
	if rl.config.IsWhitelisted(key) {
		return true
	}
	
	rl.cleanup()
	
	bucket := rl.getBucket(key)
	return bucket.Allow(1)
}

// AllowN checks if n requests are allowed for the given key
func (rl *RateLimiter) AllowN(key string, n int) bool {
	if !rl.config.Enabled {
		return true
	}
	
	if rl.config.IsWhitelisted(key) {
		return true
	}
	
	rl.cleanup()
	
	bucket := rl.getBucket(key)
	return bucket.Allow(float64(n))
}

// getBucket gets or creates a token bucket for the given key
func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	rl.mutex.RLock()
	bucket, exists := rl.buckets[key]
	rl.mutex.RUnlock()
	
	if exists {
		return bucket
	}
	
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	// Double-check after acquiring write lock
	bucket, exists = rl.buckets[key]
	if exists {
		return bucket
	}
	
	// Create new bucket
	period := rl.config.GetPeriodDuration()
	rate := float64(rl.config.Rate) / period.Seconds()
	
	bucket = &TokenBucket{
		tokens:   float64(rl.config.BurstSize),
		capacity: float64(rl.config.BurstSize),
		rate:     rate,
		lastFill: time.Now(),
	}
	
	rl.buckets[key] = bucket
	return bucket
}

// cleanup removes old buckets to prevent memory leak
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	if now.Before(rl.cleanupAt) {
		return
	}
	
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	// Remove buckets that haven't been used for an hour
	cutoff := now.Add(-1 * time.Hour)
	for key, bucket := range rl.buckets {
		bucket.mutex.Lock()
		if bucket.lastFill.Before(cutoff) {
			delete(rl.buckets, key)
		}
		bucket.mutex.Unlock()
	}
	
	rl.cleanupAt = now.Add(1 * time.Hour)
}

// Allow checks if n tokens are available
func (tb *TokenBucket) Allow(n float64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	now := time.Now()
	tb.fill(now)
	
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	
	return false
}

// fill refills the bucket based on time elapsed
func (tb *TokenBucket) fill(now time.Time) {
	elapsed := now.Sub(tb.lastFill).Seconds()
	tokensToAdd := elapsed * tb.rate
	
	tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
	tb.lastFill = now
}

// GetStats returns statistics about the rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	return map[string]interface{}{
		"enabled":      rl.config.Enabled,
		"rate":         rl.config.Rate,
		"period":       rl.config.Period,
		"burst_size":   rl.config.BurstSize,
		"key_type":     rl.config.KeyType,
		"bucket_count": len(rl.buckets),
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}