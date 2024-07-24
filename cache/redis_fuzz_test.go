package cache

import (
	"context"
	"testing"
	"time"

	"github.com/zendesk/go-generics/internal/test"
)

func Fuzz_Redis_String(f *testing.F) {
	f.Add("", "123")
	f.Add("df31", "")
	f.Add("dfadfadfadf", "2")
	f.Add("121212", "3adfa")

	f.Fuzz(func(t *testing.T, value string, updatedValue string) {
		mockRedis := &mockClient{}
		redisCache := NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

		key := "key"
		err := redisCache.Set(key, value)
		test.CheckErr(err, "Failed to set key", t)

		gotValue, ok, err := redisCache.Get(key)
		test.CheckOk(ok, "Unexpeced OK (1)", t)
		test.CheckErr(err, "Failed to get", t)
		test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
		test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
		test.CheckEqual(gotValue, "Got Value", value, t)

		err = redisCache.Set(key, updatedValue)
		test.CheckErr(err, "Failed to update", t)

		gotValue, ok, err = redisCache.Get(key)
		test.CheckErr(err, "Failed to get (2)", t)
		test.CheckEqual(gotValue, "Updated Value", updatedValue, t)
	})
}

func Fuzz_Redis_Bytes(f *testing.F) {
	f.Add([]byte{1}, []byte{})
	f.Add([]byte{1}, []byte{2})

	f.Fuzz(func(t *testing.T, value []byte, updatedValue []byte) {
		mockRedis := &mockClient{}
		redisCache := NewRedisCache[string, []byte](context.Background(), mockRedis, 10*time.Second)

		key := "key"
		gotValue, ok, err := redisCache.Get(ExpectNoResult)
		test.CheckNotOk(ok, "Unexpected ok (1)", t)
		test.CheckErr(err, "No err expected", t)

		err = redisCache.Set(key, value)
		test.CheckErr(err, "Failed to set key", t)

		gotValue, ok, err = redisCache.Get(key)
		test.CheckErr(err, "Failed to get", t)
		test.CheckOk(ok, "Expected ok (1)", t)
		test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
		test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
		test.CheckEqualEquateEmpty(gotValue, "Got Value", value, t)

		err = redisCache.Set(key, updatedValue)
		test.CheckErr(err, "Failed to update", t)

		gotValue, ok, err = redisCache.Get(key)
		test.CheckErr(err, "Failed to get (2)", t)
		test.CheckOk(ok, "Expected ok (2)", t)
		test.CheckEqualEquateEmpty(gotValue, "Updated Value", updatedValue, t)
	})
}

func Fuzz_Redis_Struct(f *testing.F) {
	type Foo struct {
		Value string
	}

	f.Add("", "")
	f.Add("faldkfjasdfk", "abc23")
	f.Add("\\x80", "0")

	f.Fuzz(func(t *testing.T, value string, updatedValue string) {
		mockRedis := &mockClient{}
		redisCache := NewRedisCache[string, Foo](context.Background(), mockRedis, 10*time.Second)

		firstValue := Foo{Value: value}

		key := "key"
		err := redisCache.Set(key, firstValue)
		test.CheckErr(err, "Failed to set key", t)

		gotValue, _, err := redisCache.Get(key)
		test.CheckErr(err, "Failed to get", t)
		test.CheckEqual(mockRedis.getKey, "Get key does not match", hashAny(key), t)
		test.CheckEqual(mockRedis.setKey, "SET key does not match", hashAny(key), t)
		if gotValue.Value != firstValue.Value {
			t.Logf("%+v : %+v", []byte(gotValue.Value), []byte(firstValue.Value))
			t.Fatal("VALUES ARE NOT THE SAME?!?")
		}
		test.CheckEqualEquateEmpty(gotValue, "Got Value", firstValue, t)

		secondValue := Foo{Value: updatedValue}
		err = redisCache.Set(key, secondValue)
		test.CheckErr(err, "Failed to update", t)

		gotValue, _, err = redisCache.Get(key)
		test.CheckErr(err, "Failed to get (2)", t)
		test.CheckEqualEquateEmpty(gotValue, "Updated Value", secondValue, t)
	})
}
