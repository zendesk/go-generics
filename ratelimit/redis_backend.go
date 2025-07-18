package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RedisClient is implemented by *redis.Client
type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd

	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
}

type RedisRateLimiterBackend struct {
	rate          int
	burstCapacity int
	rateTimeframe time.Duration
	failGateType  GateType
	limiter       *redis_rate.Limiter
	rateLock      sync.RWMutex
}

// NewRedisRateLimiterBackend creates a new rate limiter with a redis backend
//
//	rate: The rate at which the client can consume rate. If set to -1, the rate limiter will allow all requests.
//	overTime: The duration over which the rate is calculated. For instance, if you want to provide a client 1 request per second, you would set rate=1, overTime=1s.
//	burstCapacity: total burst capacity of the limiter. This must be >= `rate`. If set to less than rate, it will be set to rate (which means this bucket cannot burst beyond the throughput defined by rate)
//	for instance, if you want to provide a client 1 request per second, with a max burst of 5 requests, you would set rate=1, overTime=1s, and burstCapacity=5.
func NewRedisRateLimiterBackend(rate int, overTime time.Duration, burstCapacity int, client RedisClient) *RedisRateLimiterBackend {
	limiter := redis_rate.NewLimiter(client)

	backend := &RedisRateLimiterBackend{
		limiter: limiter,
	}

	backend.SetThroughput(rate, overTime, burstCapacity)
	return backend
}

// GetRate Returns T/F if there is rate available for execution, decrements if return is true
func (rrl *RedisRateLimiterBackend) GetRate(ctx context.Context, clientID string) (bool, error) {
	result, err := rrl.limiter.Allow(ctx, clientID, redis_rate.Limit{
		Rate:   rrl.getRate(),
		Burst:  rrl.getBurstCapacity(),
		Period: rrl.getRateTimeFrame(),
	})
	if err != nil {
		return false, err
	}

	return result.Allowed >= 1, nil
}

func (rrl *RedisRateLimiterBackend) getBurstCapacity() int {
	rrl.rateLock.RLock()
	defer rrl.rateLock.RUnlock()
	return rrl.burstCapacity
}

func (rrl *RedisRateLimiterBackend) getRate() int {
	rrl.rateLock.RLock()
	defer rrl.rateLock.RUnlock()
	return rrl.rate
}

func (rrl *RedisRateLimiterBackend) getRateTimeFrame() time.Duration {
	rrl.rateLock.RLock()
	defer rrl.rateLock.RUnlock()
	return rrl.rateTimeframe
}

// SetThroughput returns the sets the current configured rate, timeframe, and burst capacity
func (rrl *RedisRateLimiterBackend) SetThroughput(rate int, overTime time.Duration, burstCapacity int) {
	rrl.rateLock.Lock()
	defer rrl.rateLock.Unlock()

	if rate == UnlimitedRate {
		rate = MaxInt
	}

	if burstCapacity < rate {
		burstCapacity = rate
	}

	rrl.rate = rate
	rrl.rateTimeframe = overTime
	rrl.burstCapacity = burstCapacity
}

// GetThroughput returns the current configured rate, timeframe, and burst capacity
func (rrl *RedisRateLimiterBackend) GetThroughput() (int, time.Duration, int) {
	rrl.rateLock.RLock()
	defer rrl.rateLock.RUnlock()
	return rrl.rate, rrl.rateTimeframe, rrl.burstCapacity
}
