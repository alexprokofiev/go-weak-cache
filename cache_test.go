package goweakcache

import (
	"runtime/debug"
	"testing"
)

func TestCache_FailsToLoadNonExistentKey(t *testing.T) {
	t.Parallel()

	cache := New[int, int]()

	val, ok := cache.Load(1)
	if ok {
		t.Errorf("expected ok to be false, got %v", ok)
	}

	if val != nil {
		t.Errorf("expected val to be nil, got %v", val)
	}
}

func TestCache(t *testing.T) {
	t.Parallel()

	cache := New[int, int]()

	val := 1

	debug.SetGCPercent(-1)

	cache.Store(1, &val)

	valptr, ok := cache.Load(1)
	if !ok {
		t.Errorf("expected ok to be true, got %v", ok)
	}

	if valptr == nil {
		t.Errorf("expected val to be 1, got %v", val)
	}
}
