package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zendesk/go-generics/internal/test"
)

const (
	ExpectNoResult string = "emptyGet"
	ReturnErrOnSet string = "setErr"
)

type mockClient struct {
	getKey   string
	setValue []byte
	setKey   string
}

func NewMockClient() *mockClient {
	return &mockClient{}
}

func (mc *mockClient) Get(ctx context.Context, key string) *redis.StringCmd {
	mc.getKey = key
	val := redis.NewStringCmd(ctx)

	if key == hashAny(ExpectNoResult) {
		val.SetErr(redis.Nil)
	} else {
		val.SetVal(string(mc.setValue))
	}
	return val
}

func (mc *mockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	mc.setKey = key
	cmd := redis.NewStatusCmd(ctx)
	if key == hashAny(ReturnErrOnSet) {
		cmd.SetErr(errors.New("SOME NEW ERROR on cache set!"))
	}
	mc.setValue = value.([]byte)
	return cmd
}

func (mc *mockClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return nil

}

func (mc *mockClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	return nil
}

func Test_Redis_StrStr(t *testing.T) {
	mockRedis := &mockClient{}

	redisCache := NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	key := "key"
	val := "MY TES TSTRING"
	err := redisCache.Set(key, val)
	test.CheckErr(err, "Failed to set key", t)

	gotValue, _, err := redisCache.Get(key)
	test.CheckErr(err, "Failed to get", t)
	test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
	test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
	test.CheckEqual(gotValue, "Got Value", val, t)

	newValue := "adsfasdfsadfsdafasfadsfasdfa4351251341dfsdafa2341"
	err = redisCache.Set(key, newValue)
	test.CheckErr(err, "Failed to update", t)

	gotValue, _, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get (2)", t)
	test.CheckEqual(gotValue, "Got Value (2)", newValue, t)
}

func Test_Redis_Foo(t *testing.T) {
	mockRedis := &mockClient{}
	type Foo struct {
		Value string
	}

	redisCache := NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second)

	key := "key"
	val := Foo{Value: "Mdlfkajlfkjaksfsdaf"}
	err := redisCache.Set(key, val)
	test.CheckErr(err, "Failed to set key", t)

	gotValue, _, err := redisCache.Get(key)
	test.CheckErr(err, "Failed to get", t)
	test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
	test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
	test.CheckEqual(gotValue, "Got Value", val, t)

	newValue := Foo{Value: "Ofdsfjafjlfjasf"}
	err = redisCache.Set(key, newValue)
	test.CheckErr(err, "Failed to update", t)

	gotValue, _, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get (2)", t)
	test.CheckEqual(gotValue, "Got Value (2)", newValue, t)

	gotValue, _, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get (2)", t)
	test.CheckEqual(gotValue, "Got Value (2)", newValue, t)
}

func Test_Redis_GetOrSet(t *testing.T) {
	mockRedis := &mockClient{}
	type Foo struct {
		Value string
	}

	redisCache := NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second)

	key := "key"
	val := Foo{Value: "Mdlfkajlfkjaksfsdaf"}
	getOrSet := func() (Foo, error) {
		return val, nil
	}

	gotValue, fromCache, err := redisCache.GetOrSet(key, getOrSet)
	test.CheckErr(err, "Failed to get", t)
	test.CheckNotOk(fromCache, "Was from cache but shouldn't have been", t)
	test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
	test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
	test.CheckEqual(gotValue, "Got Value", val, t)

	gotValue, fromCache, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get", t)
	test.CheckOk(fromCache, "Was not from cache but should have been", t)
	test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
	test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
	test.CheckEqual(gotValue, "Got Value", val, t)

	newValue := Foo{Value: "Ofdsfjafjlfjasf"}
	err = redisCache.Set(key, newValue)
	test.CheckErr(err, "Failed to update", t)

	gotValue, _, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get (2)", t)
	test.CheckEqual(gotValue, "Got Value (2)", newValue, t)

	gotValue, _, err = redisCache.Get(key)
	test.CheckErr(err, "Failed to get (2)", t)
	test.CheckEqual(gotValue, "Got Value (2)", newValue, t)
}

func Test_Redis_GetOrSet_Error(t *testing.T) {
	mockRedis := &mockClient{}
	type Foo struct {
		Value string
	}

	redisCache := NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second)

	getOrSet := func() (Foo, error) {
		return Foo{}, fmt.Errorf("ERR")
	}

	gotValue, fromCache, err := redisCache.GetOrSet(ExpectNoResult, getOrSet)
	test.ExpectErr(err, "Failed to get", t)
	test.CheckNotOk(fromCache, "Was from cache but shouldn't have been", t)
	test.CheckEqual(Foo{}, "Got Value", gotValue, t)

	gotValue, fromCache, err = redisCache.Get(ExpectNoResult)
	test.CheckErr(err, "Failed to get", t)
	test.CheckNotOk(fromCache, "Was from cache but shouldn't have been", t)
}

func Test_Redis_GetOrSet_IgnoreSetError(t *testing.T) {
	mockRedis := &mockClient{}
	type Foo struct {
		Value string
	}

	redisCache := NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second)

	getOrSet := func() (Foo, error) {
		return Foo{}, nil
	}

	// Validate correct error type is returned
	gotValue, fromCache, err := redisCache.GetOrSet(ReturnErrOnSet, getOrSet)
	test.ExpectErr(err, "Expected err on GetOrSet but did not get one", t)
	test.CheckNotOk(fromCache, "Was from cache but shouldn't have been", t)
	test.CheckEqual(Foo{}, "Got Value", gotValue, t)

	// Validate no error is returned with IgnoreCacheSetErrors
	mockRedis = &mockClient{}
	redisCache = NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second, IgnoreCacheSetErrors[string, Foo]())

	gotValue, fromCache, err = redisCache.GetOrSet(ReturnErrOnSet, getOrSet)
	test.CheckErr(err, "Did not expect error!", t)
	test.CheckNotOk(fromCache, "Was from cache but shouldn't have been", t)
	test.CheckEqual(Foo{}, "Got Value", gotValue, t)
}
