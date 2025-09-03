package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	UnlimitedRate = -1
	MaxUint       = ^uint(0)
	MaxInt        = int(MaxUint >> 1)
)

type RateLimitedClient struct {
	limiter *rate.Limiter
}

type RateLimitedClients map[string]*RateLimitedClient

type MemoryRateLimiterBackend struct {
	rate          int
	rateTimeframe time.Duration
	burstCapacity int
	clients       RateLimitedClients
	clientLock    sync.RWMutex
	rateLock      sync.RWMutex
}

// NewMemoryRateLimiterBackend creates a new rate limiter backend in memory using Go's built-in rate limiter.
// Each client gets its own rate limiter instance with individually maintained rates.
// Parameters:
//
//	rate: The rate at which the client can consume rate. If set to -1, the rate limiter will allow all requests.
//	overTime: The duration over which the rate is calculated. For instance, if you want to provide a client 1 request per second, you would set rate=1, overTime=1s.
//	burstCapacity: total burst capacity of the limiter. This must be >= `rate`. If set to less than rate, it will be set to rate (which means this bucket cannot burst beyond the throughput defined by rate)
//	for instance, if you want to provide a client 1 request per second, with a max burst of 5 requests, you would set rate=1, overTime=1s, and burstCapacity=5.
func NewMemoryRateLimiterBackend(rateLimit int, overTime time.Duration, burstCapacity int) *MemoryRateLimiterBackend {
	clientRateData := RateLimitedClients{}
	limiter := &MemoryRateLimiterBackend{
		clients: clientRateData,
	}

	limiter.SetThroughput(rateLimit, overTime, burstCapacity)

	return limiter
}

func (rl *MemoryRateLimiterBackend) getClient(clientID string) *RateLimitedClient {
	rl.clientLock.RLock()
	client, ok := rl.clients[clientID]
	if ok {
		defer rl.clientLock.RUnlock()
		return client
	}
	rl.clientLock.RUnlock()
	rl.clientLock.Lock()
	defer rl.clientLock.Unlock()

	// Now with write lock, double check the client was not just inserted between RUnlock() and Lock(), if it was, return it
	if client, ok = rl.clients[clientID]; ok {
		return client
	}

	// if not found, initialize client with a new rate limiter
	rateLimit := rl.getRate()
	burstCapacity := rl.getBurstCapacity()

	// Handle unlimited rate case
	if rateLimit == UnlimitedRate {
		client = &RateLimitedClient{
			limiter: rate.NewLimiter(rate.Inf, burstCapacity),
		}
	} else {
		// Convert rate per timeframe to rate per second
		rateTimeframe := rl.getRateTimeFrame()
		ratePerSecond := rate.Limit(float64(rateLimit) / rateTimeframe.Seconds())
		limiter := rate.NewLimiter(ratePerSecond, burstCapacity)

		// Start with the initial rate (not full burst capacity) to match original behavior
		// The original implementation started clients with a full bucket of `rate`, not `burstCapacity`
		tokensToRemove := burstCapacity - rateLimit
		if tokensToRemove > 0 {
			limiter.AllowN(time.Now(), tokensToRemove)
		}

		client = &RateLimitedClient{
			limiter: limiter,
		}
	}

	rl.clients[clientID] = client
	return client
}

// GetRate Returns T/F if there is rate available for execution, decrements if return is true
func (rl *MemoryRateLimiterBackend) GetRate(ctx context.Context, clientID string) (bool, error) {
	if rl.getRate() <= UnlimitedRate {
		return true, nil
	}

	client := rl.getClient(clientID)
	return client.limiter.Allow(), nil
}

func (rl *MemoryRateLimiterBackend) getBurstCapacity() int {
	rl.rateLock.RLock()
	defer rl.rateLock.RUnlock()
	return rl.burstCapacity
}

func (rl *MemoryRateLimiterBackend) getRate() int {
	rl.rateLock.RLock()
	defer rl.rateLock.RUnlock()
	return rl.rate
}

func (rl *MemoryRateLimiterBackend) getRateTimeFrame() time.Duration {
	rl.rateLock.RLock()
	defer rl.rateLock.RUnlock()
	return rl.rateTimeframe
}

// SetThroughput sets the current configured rate, timeframe, and burst capacity
// and updates all existing client rate limiters
func (rl *MemoryRateLimiterBackend) SetThroughput(rateLimit int, overTime time.Duration, burstCapacity int) {
	rl.rateLock.Lock()
	defer rl.rateLock.Unlock()

	if burstCapacity < rateLimit && rateLimit != UnlimitedRate {
		burstCapacity = rateLimit
	}

	rl.rate = rateLimit
	rl.rateTimeframe = overTime
	rl.burstCapacity = burstCapacity

	// Update all existing client rate limiters with new settings
	rl.clientLock.Lock()
	defer rl.clientLock.Unlock()

	for _, client := range rl.clients {
		if rateLimit == UnlimitedRate {
			client.limiter.SetLimit(rate.Inf)
		} else {
			ratePerSecond := rate.Limit(float64(rateLimit) / overTime.Seconds())
			client.limiter.SetLimit(ratePerSecond)
		}
		client.limiter.SetBurst(burstCapacity)
	}
}

// GetThroughput returns the current configured rate, timeframe, and burst capacity
func (rl *MemoryRateLimiterBackend) GetThroughput() (int, time.Duration, int) {
	rl.rateLock.RLock()
	defer rl.rateLock.RUnlock()
	return rl.rate, rl.rateTimeframe, rl.burstCapacity
}
