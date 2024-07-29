[![Tests](https://github.com/zendesk/go-generics/actions/workflows/test_release.yml/badge.svg)](https://github.com/zendesk/zendesk/actions/workflows/test.yml)

# go-generics

Contains generic functions, data structures, and utilities for go programmers, including:

- [Functions](#functions)
- [Rate](#ratelimiter) / [Concurrency Limiters](#concurrency-limiter)
- [Caches](#caching)
- [Data structures](#data-structures)
- [Succinct Serialization](#succinct-serialization)
- [Encryption](#encryption)

Brought to you by the Zendesk Lockbox team. 


## Functions

The `functions` package contains dozens of generic functions with custom options support to allow fast-mapping with, or without concurrency,
client side rate limiting, automated retries, and more.

Functions **not** prefixed with `Go` will run serially.

Functions prefixed with `Go` will run concurrently, and may be tuned with the additional options:
- `RateLimitOption`: Limits maximum iterations that may be executed over a specified timeframe
  - e.g. `functions.RateLimitOption(10, time.Second)`
- `RetryOption`: Retries a function if it returns an error with linear progressive backoff (backoff duration * retry number)
  - e.g. `functions.RetryOption(3, time.Millisecond * 500)`
- `RandomOrderOption`: The targeted function will randomly order its execution rather than iterating over elements in the provided order
- `DiscardResultIfErrOption`: Mapping functions will discard results when errors are returned
- `ConcurrencyLimitOption`: limits the concurrency of a concurrent mapping function to protect against open file limits, connection limits, etc. To run serially, set concurrency to 1.

### Comparison:
- `EqualIgnoreOrder[T comparable](slices ...[]T) bool`
- `Contains[T comparable](list []T, item T) bool`
- `ContainsAny[T comparable](A []T, B []T) bool`
- `ContainsDeepEqual[T any](list []T, item T) bool`

```go
   functions.EqualIgnoreOrder([]int{1, 2, 3}, []int{3, 2, 1}) // true
   functions.Contains([]int{1, 2, 3}, 2) // true
   functions.ContainsAny([]int{1, 2, 3}, []int{4, 5, 6}) // false
```

### Iterative
- `Each[T any](items []T, fn func(T))`
- `EachMergeErrs[T any](items []T, fn func(T) error) error`
- `GoEach[T any](items []T, fn func(T), opts ...Option)`
- `GoEachWithErrs[T any](items []T, fn func(T) error, opts ...Option) (errs []error)`
- `GoEachMapWithErrs[K comparable, V any](items map[K]V, fn func(K, V) error, opts ...Option) (errs []error)`

```go
  
  // Iterate over slice
  functions.Each([]int{1, 2, 3}, func(i int) {
    fmt.Println(i)
  })
  
  // iterate over slice, merge errors and return as a single error.
  err := functions.EachMergeErrs([]int{1, 2, 3}, func(i int) error {
    if i == 2 {
        return fmt.Errorf("Error encountered")
    }
    return nil
  })

  // Iterate over slice concurrently, with rate limiting
  errs := functions.GoEachWithErrs([]{1,2,3}, func(i int) error {
	  if i%2 == 0 {
		  return fmt.Errorf("Error encountered")
      }
	  return nil
  }, functions.RateLimitOption(1, time.Second))

  // Iterate over map  with rate limiting
  myMap := map[int]string{1: "one", 2: "two", 3: "three"}
  functions.GoEachMapWithErrs(myMap, func(k int, v string) error {
      if k == 2 {
          return fmt.Errorf("Error encountered")
      }
    return nil
  }, functions.RateLimitOption(1, time.Second))

```

### Filters
- `Find[T interface{}](from []T, filter func(T) bool) (item T, wasFound bool)`
- `Filter[T any](from []T, filter func(T) bool) []T`
- `FilterMap[K comparable, V any](from map[K]V, filter func(k K, v V) bool) map[K]V`

```go 
  // Find the first even number and return it
  firstEven, found := functions.Find([]int{1, 2, 3, 4}, func(i int) bool {
    return i%2 == 0 
  }) // 2, true

  // Find all evens and return them
  evens := functions.Filter([]int{1, 2, 3, 4}, func(i int) bool {
      return i%2 == 0 
  }) // [2, 4]    
	
  // Filter all even keys from the map
  evenValues := functions.FilterMap(map[int]string{1: "one", 2: "two", 3: "three", 4: "four"}, func(k int, v string) bool {
    return k%2 == 0
  }) // {2: "two", 4: "four"}
```


### Mapping

#### From X to Slice
- `Map[T any, Y any](from []T, converter func(T) Y) []Y`
- `MapWithErrs[T any, Y any](from []T, converter func(T) (Y, error)) ([]Y, []error)`
- `MapMergeErrs[T any, Y any](from []T, converter func(T) (Y, error)) ([]Y, error)`
- `GoMap[T any, Y any](items []T, converter func(T) Y, opts ...Option) []Y`
- `GoMapWithErrs[T any, Y any](items []T, f func(T) (Y, error), opts ...Option) (results []Y, errs []error)`
- `GoMapToMany[T any, Y any](items []T, converter func(T) []Y, opts ...Option) (results []Y)`
- `GoMapToManyWithErrs[T any, Y any](items []T, converter func(T) ([]Y, error), opts ...Option) (results []Y, errs []error)`
- `MapToSlice[K comparable, V any, Z any](from map[K]V, converter func(k K, v V) Z) []Z`
- `GoMapToSlice[K comparable, V any, Z any](items map[K]V, converter func(K, V) Z, opts ...Option) []Z`
- `GoMapToSliceWithErrs[K comparable, V any, Z any](items map[K]V, converter func(K, V) (Z, error), opts ...Option) (results []Z, errs []error)`

```go 
  // Double each value in the slice
  doubled := functions.Map([]int{1, 2, 3}, func(i int) int {
      return i * 2
  }) // [2, 4, 6]
  
  // Convert slice of IDs to []Foo via API lookup
  // Discard result if error is returned.
  // Rate limit requests to 10 per second
  results, errs := functions.GoMapWithErrs([]string{"id1", "id2", "id3"}, func(id string) (Foo, error) {
      return api.Lookup(i) // returns (Foo, error)
  }, functions.RateLimitOption(10, time.Second), functions.DiscardResultIfErrOption())
  
  // Convert slice to a larger slice where each item returns 1+ items
  results := functions.GoMapToMany([]int{1, 2, 3}, func(i int) []int {
      return []int{i, i+1}
  }) // [1, 2, 2, 3, 3, 4]
  
  // Convert map[int]string to slice []string
  results := functions.GoMapToSlice(map[int]string{1: "one", 2: "two", 3: "three"}, func(k int, v string) string {
      return fmt.Sprintf("%d: %s", k, v)
  }) // ["1: one", "2: two", "3: three"]

```


#### From X to Map
- `ToMap[T any, K comparable, V any](from []T, converter func(T) (K, V)) map[K]V`
- `GoToMap[T any, K comparable, V any](items []T, f func(T) (K, V), opts ...Option) map[K]V`
- `GoToMapWithErrs[T any, K comparable, V any](items []T, f func(T) (K, V, error), opts ...Option) (results map[K]V, errs []error)`

```go 

  // Convert []int to map[int]string
  functions.ToMap([]int{1, 2, 3}, func(i int) (int, string) {
      return i, fmt.Sprintf("string-%d", i)
  }) // {1: "string-1", 2: "string-2", 3: "string-3"}

```

### Reduce
- `Reduce[T any, Y any](from []T, to Y, reducer func(T, Y) Y) Y`

```go 
  // Reduce to sum of []int
  sum := functions.Reduce([]int{1, 2, 3}, 0, func(i int, sum int) int {
      return sum + i
  }) // 6
```

### Other
- `RunWithRetries[T any](fn func(t T) error, item T, numRetries int, backoffInterval time.Duration) error`
- `Min[T cmp.Ordered](values ...T) T`
- `Max[T cmp.Ordered](values ...T) T`
- `Copy[K comparable, V any](items map[K]V) map[K]V`
- `Convert[T any, Y any](from T, converter func(T) Y) Y`

```go 
  // Run something, and retry if you get an error
  functions.RunWithRetries(func() error {
      return api.Call()
  }, 3, time.Millisecond * 500) // Retry 3 times with 500ms progressive backoff


  // Copy a map
  copied := functions.Copy(map[int]string{1: "one", 2: "two", 3: "three"}) // {1: "one", 2: "two", 3: "three"}
```
  
### Slice
  
- `Intersection[T comparable](a, b []T) []T`
- `Dedupe[T comparable](items []T) []T`
- `DedupeByHash[T comparable](items []T, hashFn func(t T) string) []T`
- `Shuffle[T any](items []T) []T`
- `RemoveNils[T any](from []T) []T`
- `Generalize[T any](from []T) []interface{}`
- `Join[T any](items []T, separator string) string`

```go 
    // Return the intersection of two slices
    intersection := functions.Intersection([]int{1, 2, 3}, []int{2, 3, 4}) // [2, 3]

    // Dedupe a comparable slice
    uniques := functions.Dedupe([]int{1, 2, 2, 3, 3, 3}) // [1, 2, 3]

    // Shuffle as lice
    shuffled := functions.Shuffle([]int{1, 2, 3, 4, 5}) // [?, ?, ?, ?, ?]
```

**Advanced Example**
```go
// Execute an API call concurrently, one for each ID in the list, and return the result, or an error.
// Rate-limit requests to 10 per second.
// Limit max concurrent requests to 5
// If a request returns an error, it will be retried up to 3 times, with a 500 millisecond progressive backoff.

fooIds := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
foos, errs := functions.GoMapWithErrs(fooIds, func(id string) (Foo, error) {
    foo, err := fooAPI.GetFoo(id)
        if err != nil {
            return Foo{}, fmt.Errorf("Error encountered, this will trigger a retry: %w", err)
        }
    return foo, err
}, RateLimitOption(10, time.Second), RetryOption(3, time.Millisecond*500), ConcurrencyLimitOption(5))
```

## Data Structures

The `types` package contains some conspicuously missing go data structures, including:
- `Set`
  - `NewSet[T comparable](items ...T) Set[T]`
  - `NewHashSet[T any](items ...T) Set[T]` 
    - ^^ May be used with any data structure, even non-comparable ones
  - `NewHashSetWithHashFn[T](fn HashFn, items ...T)`
    - You may provide your own custom hash function. `func(t T) string`
- `Stack`

```go

  mySet := datastructures.NewSet(1, 2, 3, 4, 5, 5, 5, 5)
  // {1, 2, 3, 4, 5} (order is not guaranteed)

  // Hash set automatically dedupes non-comparable types
  type foo struct {
    Name string
    Age int
  }
  
  foos := []foo{{"James", 30}, {"Bob", 44}, {"James", 30}}   
  myHashSet := datastructures.NewHashSet(foos...)
  // {{"James", 30}, {"Bob", 44}}
	
  // Hash set with custom hash function (only dedupe by name)
  foos := []foo{{"James", 30}, {"Bob", 44}, {"James", 99}}
  myHashSet := datastructures.NewHashSetWithHashFn(func(f foo) string {
      return f.Name
  }, foos...)
  // {{"James", 30}, {"Bob", 44}} OR  {{"James", 99}, {"Bob", 44}}
	
```

### Future plans
- Add an option to enable synchronization of the datastructures to prevent concurrent modification. Right now these datastructures are not thread safe.

## Caching

The `cache` package contains a generic cache implementation that supports dynamic backends, redis, or in-memory. You may also supply your own backend. 
Additionally, a `fail-through` cache may be supplied, for instance, so the in-memory cache can be checked first, with a fail-through to redis on a
cache miss. 

### Features

- Generic implementation that supports all types.
- Time-to-live for items set in the cache may be configured
- Fail-through cache may be configured to configure multiple levels of caching. If the key is missing from the primary, the secondary will be queried
- Supports transparent encryption / decryption wrapper for configuring a encrypted cache in memory and/or redis.

Future goals / features:
- Sized based capacity
- Custom eviction (LRU, LFU, etc)
  - In memory cache uses LRU based eviction when capacity is reached.

### Cache Examples

Example 1: Simple in-memory cache

```go 
// Basic in memory cache example
type Person struct {
	Name string
	Age int
}
ttl := time.Minute
capacity := uint64(4096)
cache := cache.NewInMemoryCache[string, Person](ttl,
    cache.WithCapacity[string, string](capacity)
)

// Set a value in the cache
cache.Set(userID, Person{Name: "James", Age: 30})

// Get a value from the cache, or if it doesn't exist, look it up from the DB, set it in the cache, and return it
cache.GetOrSet(userID, func() (V, error) { 
	return db.ReadPerson(userID)
})
```

Example 2: In memory cache with redis fail-through

```go
// In Memory cache with Redis Cache fail-through
type Person struct {
    Name string
    Age int
}

ttl := time.Minute
capacity := uint64(4096)


client, _ := NewRedisClient(redisCfg)
failThrough := cache.NewRedisCache[K, Person](context.Background(), client, ttl)

cache := cache.NewInMemoryCache[string, Person](ttl,
    cache.WithCapacity[string, string](capacity),
    cache.WithFailThroughCache[string, string](failThrough),
)

// Set a value in the cache (this will also be set in the fail-through cache)
cache.Set(userID, Person{Name: "James", Age: 30})

// Get a value from the cache, or if it doesn't exist, look it up from the DB, set it in the cache, and return it
// This will also be set in the fail-through cache
cache.GetOrSet(userID, func() (V, error) {
    return db.ReadPerson(userID)
})


// Get a value from the cache, if it is found in the fail-through cache, it will be added to the primary cache as it is returned.
user, wasFound, err := cache.Get(userID)
```

## RateLimiter

A generic rate-limiter implementation is provided with support for various backends, including: redis, or in-memory. You may also supply your own backend.

### Features
- Supports "burst" via leaky bucket algorithm.
- Supports "clientID" parameter for custom rate-limiting per client
- Adds fail-open or fail-closed configuration for customization of operation in event of backend failure errors (network timeout, etc)
- Supports custom prefixes, so a single backend may serve many limiters, each with custom client sets.

### Examples

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
    writeLimiter, err := ratelimit.NewRateLimiter(ratelimit.FailOpen, backend, ratelimit.WithPrefixOption("writes"))
		
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

## Concurrency Limiter

Limits max concurrency of the Run() function based on config:

### Features
- Allows providing of onComplete callback function after the provided function completes

```go
// Limit concurrency to 5 concurrent executions. The provided Run() function will be executed in a goroutine.

limiter := ratelimit.NewConcurrencyLimiter(5)

for i := 0; i < 20; i++ {
      limiter.Run(func() {
        fmt.Printf("Run #: %d \n", i)
        time.Sleep(time.Second)
    })
}

// With on-complete callback
for i := 0; i < 20; i++ {
    limiter.Run(func() {
        fmt.Printf("Run #: %d executing \n", i)
        time.Sleep(time.Second)
    }, ratelimit.WithOnCompleteCallback(func() {
        fmt.Printf("Callback executed for: %d \n", i)
    }))
}

// Wait for callbacks before existing
time.Sleep(time.Second * 5)


```

## Succinct Serialization

Simplifies serializing / deserialization of data from various formats. Also allows dynamic type discovery and conversion for 
dynamic use cases.

### Future plans:
- Allow the client to provide custom serializer / deserialization functions via an option.

### Examples
```go
// Deserialize a JSON string into a struct
type Person struct {
    Name string `json: "name"`
    Age int     `json: "age"`
}

// From JSON to Person
input := "{\"name\": \"James\", \"age\": 30}"
person, err := serialize.NewSerializer[Person]().FromJsonString(input).ToT()

// From Person to JSON
json, err := serialize.NewSerializer[Person]().FromT(person).ToJsonString()

// From JSON to []*Person
input = "[{\"name\": \"James\", \"age\": 30}, {\"name\": \"Bob\", \"age\": 44}]"
people, err := serialize.NewSerializer[[]*Person]().FromJsonString(input).ToT()

// From Person to []byte
bytes, err := serialize.NewSerializer[any]().FromT(person).ToBytes()

// From Person to B64String
bytes, err := serialize.NewSerializer[any]().FromT(person).ToB64String()

```

In some instances the same code may need to dynamically serialize or deserialize data from or to a variable type, in these instances, you may use FromDynamicType and ToDynamicType.
If one of these methods is called, reflection or generics may be used to dynamically determine the source or target type.
```go
// Dynamic type example
type Animal struct {
    Type string `json: "type"`
    Age int     `json: "age"`
}

person := Person{Name: "James", Age: 30}

// From any to JSON
jsonStr, err := NewSerializer[any]().FromDynamicType(person).ToJsonString()

//  To dynamic type
var typ T // Animal
input := "{\"type\": \"Gorilla\", \"age\": 2}" // JSON String

// FromDynamicType: detected the input was of type JSON String,
// ToDynamicType:   serialize.Reflect indicates that we should use reflection to determine the type of "typ" and deserialize input into that type
// In this instance, if typ is a struct, `T` will be used to deserialize the input into a struct, if the `typ` is a []byte or string, 
// then normal deserialization will be used to marshal the input into that type
gorilla, err := NewSerializer[T]().FromDynamicType(input).ToDynamicType(serialize.Reflect, typ)
```

## Encryption

### Features 
- Encrypt / decrypt generic types using AES-256 encryption.

### Future enhancements
- Support for supplying your own custom encryption backend.

### Example

Encrypt with your own password + nonce.
```go
    // Encrypt / decrypt transparently with a auto-generated passwrod
    type Foo struct {
        Name string
        Age int
    }

    // Encrypt / decrypt transparently with a provided password / nonce. If you are persisting the encrypted data, use this approach.
    ed, err := NewWithPasswordNonce[Foo]([]byte("password"), []byte("my-nonce-123"), []byte("salt"), 9000)
    if err != nil {
        return err
    }

    encryptedBytes, err := ed.Encrypt(tt.input)
    if err != nil {
        return err
    }
        
    decrypted, err := ed.Decrypt(encryptedBytes)
    if err != nil {
		return err
    }
```

Encrypt with an auto-generated secure password + nonce. 
```go
	// A secure password will be generated, do not use this constructor if you intend to persist the encrypted data.
    ed, err := encryption.New[Foo]([]byte("mySalt"), tt.iterations)
    if err != nil {
        return err
    }

    encryptedBytes, err := ed.Encrypt(tt.input)
    if err != nil {
        return err
    }
        
    decrypted, err := ed.Decrypt(encryptedBytes)
    if err != nil {
		return err
    }
```