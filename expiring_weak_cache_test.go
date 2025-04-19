package goweakcache

import (
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

func TestCacheWithExpiration_KeyExpired(t *testing.T) {
	t.Parallel()

	cache := NewWithExpiration[string, string]()

	val := "world"

	cache.StoreEX("hello", &val, time.Now().Add(time.Microsecond))

	<-time.After(time.Microsecond)

	valPtr, ok := cache.Load("hello")
	if ok {
		t.Errorf("expected ok to be false, got %v", ok)
	}

	if valPtr != nil {
		t.Errorf("expected val to be nil, got %v", val)
	}
}

func TestCacheWithExpiration_KeyNotExpiredAndGC(t *testing.T) {
	t.Parallel()

	cache := NewWithExpiration[string, string]()

	val := "world"

	debug.SetGCPercent(-1)

	cache.StoreEX("hello", &val, time.Now().Add(time.Hour))

	valPtr, ok := cache.Load("hello")
	if !ok {
		t.Errorf("expected ok to be true, got %v", ok)
	}

	if valPtr == nil {
		t.Errorf("expected val not to be nil, got %v", val)
	}

	runtime.GC()

	valPtr, ok = cache.Load("hello")
	if ok {
		t.Errorf("expected ok to be false, got %v", ok)
	}

	if valPtr != nil {
		t.Errorf("expected val not be nil, got %v", val)
	}
}

func TestCacheWithExpiration_KeyDeletedAfterGC(t *testing.T) {
	t.Parallel()

	cache := NewWithExpiration[string, string]()

	val := "world"

	cache.StoreEX("hello", &val, time.Now().Add(time.Microsecond))

	<-time.After(time.Microsecond)

	valPtr, ok := cache.Load("hello")
	if ok {
		t.Errorf("expected ok to be false, got %v", ok)
	}

	if valPtr != nil {
		t.Errorf("expected val to be nil, got %v", valPtr)
	}

	runtime.GC()

	cache.weakCache.data.Range(
		func(key string, value cacheEntry[expirableCacheEntry[string]]) bool {
			t.Errorf("expected that cache key %q to be deleted, value: %+v", key, value)

			return true
		},
	)
}
