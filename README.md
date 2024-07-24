# go-generics

Contains generic functions, data structures, and utilities for go programmers, including:

- Functions
- Rate / Concurrecy Limiters
- Generic Caches 
- Data structures

Brought to you by the Zendesk Lockbox team. 

## Functions

The `functions` package contains dozens of generic functions with custom options support to allow fast-mapping with, or without concurrency,
client side rate limiting, automated retries, and more.


Functions **not** prefixed with Go will run serially, and may be tuned with the below options:
- RateLimitOption: Limits maximum iterations that may be executed over a specified timeframe
    - e.g. functions.RandomOrderOption(10, time.Second)
- RetryOption: Retries a function if it returns an error with progressive backoff
    - e.g. functions.RetryOption(3, time.Millisecond * 500)
- RandomOrderOption: The targeted function will randomly order its execution rather than iterating over elements in the provided order
- DiscardResultIfErrOption: Mapping functions will discard results when errors are returned

Functions prefixed with `Go` will run concurrently, and may be tuned with the additional options:
- ConcurrencyLimitOption: limits the concurrency of a concurrent mapping function to protect against open file limits, connection limits, etc.

#### Functions Examples

```go
// Execute an API call concurrently, one for each ID in the list, and return the result, or an error.
// Rate-limit requests to 10 per second.
// If a request returns an error, it will be retried up to 3 times, with a 500 millisecond progressive backoff.
// So the first retry will back off by 500ms, the second by 1 second, the third by 1.5 seconds

fooIds := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
foos, errs := GoMapWithErrs(fooIds, func(id string) (Foo, error) {
    foo, err := fooAPI.GetFoo(id)
        if err != nil {
            return Foo{}, fmt.Errorf("Error encountered, this will trigger a retry: %w", err)
        }
    return foo, err
}, RateLimitOption(10, time.Second), RetryOption(3, time.Millisecond*500))
```


## Caching
 Todo: Fill in cache documentation




## RateLimiter

A generic rate-limiter implementation is provided, that supports dynamic backends, redis, or in-memory. You may also supply your own backend.

### Features
- Supports "Burst" via leaky bucket algorithm.
- Supports "clientID" parameter for custom rate-limiting per client
- Adds fail-open or fail-closed configuration for customization of operation in event of backend failure errors (network timeout, etc)
- Supports custom prefixes, so a single backend may serve many custom limiters, each with custom client sets.

#### RateLimiter Examples

Example 1: Rate limit by client, and return if rate has been exceeded
```go
// Creates a new rate limit that will limit each client to 1 request per second with an allowable max burst of 5 req/sec
type myServer struct {
      limiter *ratelimit.RateLimiter
}

func NewMyServer() *myServer {
      backend := ratelimit.NewMemoryRateLimiterBackend(1, time.Second, 5)
      limiter := ratelimit.NewRateLimiter(ratelimit.FailClosed, backend)
      return &myHandler{limiter: limiter}
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
      clientID := r.Header.Get("X-Client-ID")
      if !limiter.GetRateForClient(clientID) {
            w.WriteHeader(http.StatusTooManyRequests)
            return
      }
	  
	  // Serve request, available rate for client was decremented
}
```

Multiple rate limiters with shared redis backend and different rate limits
```go
    client, _ := NewRedisClient(redisConfig)
	  
    // Both backends connect to the same redis instance
    readBackend = ratelimit.NewRedisRateLimiterBackend(readsPerSecond, time.Second, readBurstCapacity, client)
    writeBackend = ratelimit.NewRedisRateLimiterBackend(writesPerSecond, time.Second, writeBurstCapacity, client)
    
    // if redis is unavailable, rate limiter fails OPEN, all reads will be accepted
	readLimiter, err := ratelimit.NewRateLimiter(ratelimit.FailClosed, backend, ratelimit.WithPrefixOption("reads"))
	

	// if redis is unavailable, rate limiter fails CLOSED, and will not allow any writes
    writeLimiter, err := ratelimit.NewRateLimiter(ratelimit.FailClosed, backend, ratelimit.WithPrefixOption("writes"))
		
	// If we want to dynamically adjust throughput for writes, we can
	writeLimiter.SetThroughput(20, time.Second, 50)
    	
	// If we want to block and wait for available rate
	_ = readLimiter.WaitForRateForClient(ctx, clientID)
	doSomething()
	
	// If we want to wait for available rate, but bail if no rate is available after 5 seconds
	hasRate := readLimiter.WaitForRateWithTimeoutForClient(ctx, clientID, time.Second * 5)
	if !hasRate {
        return fmt.Errorf("Rate limit exceeded")
    }
	doSomething()
}
```
