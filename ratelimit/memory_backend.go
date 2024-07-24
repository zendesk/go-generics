package ratelimit

import (
	"context"
	"sync"
	"time"
)

const (
	UnlimitedRate = -1
	MaxUint       = ^uint(0)
	MaxInt        = int(MaxUint >> 1)
)

type RateLimitedClient struct {
	rateBucket         int
	rateBucketLock     sync.RWMutex
	lastRateUpdateLock sync.RWMutex
	lastRateUpdate     time.Time
}

type RateLimitedClients map[string]*RateLimitedClient

type MemoryRateLimiterBackend struct {
	rate          int
	rateTimeframe time.Duration
	burstCapacity int
	completeChan  chan bool
	closeOnce     sync.Once
	clients       RateLimitedClients
	clientLock    sync.RWMutex
	rateLock      sync.RWMutex
}

// NewMemoryRateLimiterBackend creates a new rate limiter backend in memory. A goroutine is spawned that will maintain the bucket for all added clients. The amount of rate in the
// bucket cannot ever exceed the provided rate over duration. Clients have individually maintained rates.
// Parameters:
//
//	rate: The rate at which the client can consume rate. If set to -1, the rate limiter will allow all requests.
//	overTime: The duration over which the rate is calculated. For instance, if you want to provide a client 1 request per second, you would set rate=1, overTime=1s.
//	burstCapacity: total burst capacity of the limiter. This must be >= `rate`. If set to less than rate, it will be set to rate (which means this bucket cannot burst beyond the throughput defined by rate)
//	for instance, if you want to provide a client 1 request per second, with a max burst of 5 requests, you would set rate=1, overTime=1s, and burstCapacity=5.
func NewMemoryRateLimiterBackend(rate int, overTime time.Duration, burstCapacity int) *MemoryRateLimiterBackend {
	clientRateData := RateLimitedClients{}
	limiter := &MemoryRateLimiterBackend{
		clients:      clientRateData,
		completeChan: make(chan bool, 1),
	}

	limiter.SetThroughput(rate, overTime, burstCapacity)

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

	// Now with write lock, double check the client was not just inserted between RUnlock() and Lock(), if it was, return it
	if client, ok = rl.clients[clientID]; ok {
		defer rl.clientLock.Unlock()
		return client
	}

	// if not found, initialize client with a full bucket of rate
	client = &RateLimitedClient{
		rateBucket:     rl.getRate(),
		lastRateUpdate: time.Now(),
	}

	defer rl.clientLock.Unlock()
	rl.clients[clientID] = client

	return client
}

// hasRate returns true if there is rate available on the specified client. This does not consume rate
func (rl *MemoryRateLimiterBackend) hasRate(clientID string) bool {
	client := rl.getClient(clientID)
	client.rateBucketLock.RLock()

	defer client.rateBucketLock.RUnlock()
	return client.rateBucket > 0
}

// GetRate Returns T/F if there is rate available for execution, decrements if return is true
func (rl *MemoryRateLimiterBackend) GetRate(ctx context.Context, clientID string) (bool, error) {
	if rl.getRate() <= UnlimitedRate {
		return true, nil
	}

	// if no rate is available, check for new rate to be added
	if !rl.hasRate(clientID) {
		hasRate := rl.maintainRateBucket(clientID)
		if !hasRate {
			return false, nil
		}
	}

	return rl.consumeRate(ctx, clientID), nil
}

// consumeRate removes a single rate from the bucket. Returns false if consume was unsuccessful because someone else got it
func (rl *MemoryRateLimiterBackend) consumeRate(_ context.Context, clientID string) bool {
	client := rl.getClient(clientID)
	client.rateBucketLock.Lock()
	defer client.rateBucketLock.Unlock()
	if client.rateBucket > 0 {
		client.rateBucket--
		return true
	}
	return false
}

// MaintainRateBucket will add rate to a client's bucket if they should have more rate. If rate is added, true is returned.
func (rl *MemoryRateLimiterBackend) maintainRateBucket(clientID string) bool {
	client := rl.getClient(clientID)
	// Refresh bucket with tokens if it's time
	client.lastRateUpdateLock.Lock()
	timeForUpdate := time.Since(client.lastRateUpdate) >= rl.getRateTimeFrame()
	defer client.lastRateUpdateLock.Unlock()
	if timeForUpdate {
		// Calculate how much rate to add, you may not add more than the BurstCapacity
		rateToAdd := min(rl.getRate()*int(time.Since(client.lastRateUpdate)/rl.getRateTimeFrame()), rl.getBurstCapacity())
		client.rateBucketLock.Lock()
		client.lastRateUpdate = time.Now()
		client.rateBucket = client.rateBucket + rateToAdd
		client.rateBucketLock.Unlock()
		return true
	}

	return false
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

// SetThroughput returns the sets the current configured rate, timeframe, and burst capacity
func (rl *MemoryRateLimiterBackend) SetThroughput(rate int, overTime time.Duration, burstCapacity int) {
	rl.rateLock.Lock()
	defer rl.rateLock.Unlock()

	if rate == UnlimitedRate {
		rate = MaxInt
	}

	if burstCapacity < rate {
		burstCapacity = rate
	}

	rl.rate = rate
	rl.rateTimeframe = overTime
	rl.burstCapacity = burstCapacity
}

// GetThroughput returns the current configured rate, timeframe, and burst capacity
func (rl *MemoryRateLimiterBackend) GetThroughput() (int, time.Duration, int) {
	rl.rateLock.RLock()
	defer rl.rateLock.RUnlock()
	return rl.rate, rl.rateTimeframe, rl.burstCapacity
}
