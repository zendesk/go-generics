package cache_test

import (
	"testing"
	"time"

	"github.com/zendesk/go-generics/cache"
	"github.com/zendesk/go-generics/cache/internal/test"
)

type MockCache[K comparable, V any] struct {
	WasGot     []K
	WasSet     []K
	WasDeleted []K
	WasPurged  bool
}

func (m *MockCache[K, V]) Get(key K) (V, bool, error) {
	var v V
	m.WasGot = append(m.WasGot, key)
	return v, true, nil
}

func (m *MockCache[K, V]) Set(key K, _ V) error {
	m.WasSet = append(m.WasSet, key)
	return nil
}

func (m *MockCache[K, V]) Delete(key K) error {
	m.WasDeleted = append(m.WasDeleted, key)
	return nil
}

func (m *MockCache[K, V]) Purge() error {
	m.WasPurged = true
	return nil
}

func (m *MockCache[K, V]) GetOrSet(key K, orSet func() (V, error)) (val V, wasFoundInCache bool, err error) {
	v, err := orSet()
	return v, false, err
}

func Test_InMemory_WithFailThrough(t *testing.T) {
	failThrough := &MockCache[string, string]{}
	cache := cache.NewInMemoryCache[string, string](time.Second*1,
		cache.WithFailThroughCache[string, string](failThrough),
		cache.WithCapacity[string, string](uint64(100)))

	// Set item
	err := cache.Set("item1", "foo")
	test.CheckErr(err, "Failed to set item1", t)

	// Verify item was set in backup cache
	test.CheckContains(failThrough.WasSet, "item1", t)

	// Sleep, call Get (after expiration of main cache), verify call to get on fail-through
	time.Sleep(time.Second + time.Millisecond*100)
	_, _, err = cache.Get("item1")
	test.CheckErr(err, "Got error getting item", t)
	test.CheckContains(failThrough.WasGot, "item1", t)

	// Get again -- this should hit main cache again (not failthrough), b/c it should be refreshed from the prior GET
	_, ok, err := cache.Get("item1")
	test.CheckOk(ok, "OK expected but was not true", t)
	test.CheckErr(err, "Got error getting item", t)
	test.CheckContains(failThrough.WasGot, "item1", t)
	test.CheckOk(len(failThrough.WasGot) == 1, "Got should have hit fail-through cache only 1 time", t)

	// Get again -- this should hit main cache again (not failthrough), b/c it should be refreshed from the prior GET
	_, _, err = cache.Get("item1")
	test.CheckErr(err, "Got error getting item", t)
	test.CheckContains(failThrough.WasGot, "item1", t)
	test.CheckOk(len(failThrough.WasGot) == 1, "Got should have hit fail-through cache only 1 time", t)

	time.Sleep(time.Second)
	// Get again -- this should hit fail-through due to expiration of main again.
	_, ok, err = cache.Get("item1")
	test.CheckErr(err, "Got error getting item", t)
	test.CheckContains(failThrough.WasGot, "item1", t)
	test.CheckOk(ok, "OK expected but was not true (2)", t)
	test.CheckOk(len(failThrough.WasGot) == 2, "Got should have hit fail-through cache only 1 time", t)

	// Call delete
	err = cache.Delete("item1")
	test.CheckErr(err, "Unexpected err on delete", t)
	test.CheckContains(failThrough.WasDeleted, "item1", t)

	// Purge
	err = cache.Purge()
	test.CheckErr(err, "Unexpected err on purge", t)
	test.CheckOk(failThrough.WasPurged, "Was not purged!", t)

	// Verify set was only called once
	test.CheckOk(len(failThrough.WasSet) == 1, "Set should have only called to fail-through once, on the explicit SET at the start", t)
}

func Test_InMemory_WithFailThroughMiss(t *testing.T) {
	failThrough := cache.NewInMemoryCache[string, string](time.Second * 1)
	cash := cache.NewInMemoryCache[string, string](time.Second*1,
		cache.WithFailThroughCache[string, string](failThrough),
		cache.WithCapacity[string, string](uint64(100)))

	item, found, err := cash.Get("MISSING")
	test.CheckNotOk(found, "Found not expected", t)
	test.CheckErr(err, "No err expected", t)
	test.CheckEqual(item, "EMpty string expected", "", t)

	item, found, err = cash.Get("MISSING")
	test.CheckNotOk(found, "Found not expected (2)", t)
	test.CheckErr(err, "No err expected (2) ", t)
	test.CheckEqual(item, "EMpty string expected", "", t)
}

func Test_InMemory_Zero_TTL(t *testing.T) {
	// if no TTL is provided, this cache should not cache anything.
	cash := cache.NewInMemoryCache[string, string](0,
		cache.WithCapacity[string, string](uint64(100)))

	err := cash.Set("item1", "foo")
	test.CheckErr(err, "Unexpecgte derr setting item", t)
	item, found, err := cash.Get("item1")
	test.CheckNotOk(found, "Found not expected", t)
	test.CheckErr(err, "No err expected", t)
	test.CheckEqual(item, "Empty string expected", "", t)

	item, found, err = cash.GetOrSet("item1", func() (string, error) {
		return "foo", nil
	})
	test.CheckNotOk(found, "Found not expected", t)
	test.CheckErr(err, "No err expected", t)
	test.CheckEqual(item, "Foo expected", "foo", t)

}

func Test_InMemory_Zero_Size(t *testing.T) {
	// if size is 0, this should not cache anything.
	cash := cache.NewInMemoryCache[string, string](time.Second,
		cache.WithCapacity[string, string](uint64(0)))
	err := cash.Set("item1", "foo")
	test.CheckErr(err, "Unexpected err setting item", t)
	item, found, err := cash.Get("item1")
	test.CheckNotOk(found, "Found not expected", t)
	test.CheckErr(err, "No err expected", t)
	test.CheckEqual(item, "EMpty string expected", "", t)

	item, found, err = cash.GetOrSet("item1", func() (string, error) {
		return "foo", nil
	})
	test.CheckNotOk(found, "Found not expected", t)
	test.CheckErr(err, "No err expected", t)
	test.CheckEqual(item, "Foo expected", "foo", t)

}
