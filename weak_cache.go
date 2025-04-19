package goweakcache

import (
	"runtime"
	"weak"

	"github.com/puzpuzpuz/xsync/v4"
)

type ComputeOp = xsync.ComputeOp

const (
	CancelOp = xsync.CancelOp
	UpdateOp = xsync.UpdateOp
	DeleteOp = xsync.DeleteOp
)

type (
	WeakCache[K comparable, V any] struct {
		data *xsync.Map[K, cacheEntry[V]]
	}

	cacheEntry[V any] weak.Pointer[V]
)

func New[K comparable, V any]() *WeakCache[K, V] {
	return &WeakCache[K, V]{
		xsync.NewMap[K, cacheEntry[V]](),
	}
}

func (c *WeakCache[K, V]) Load(key K) (*V, bool) {
	var val *V

	entry, ok := c.data.Load(key)
	if ok {
		val = weak.Pointer[V](entry).Value()

		ok = val != nil
	}

	return val, ok
}

func (c *WeakCache[K, V]) Delete(key K) {
	c.data.Delete(key)
}

func (c *WeakCache[K, V]) Clear() {
	c.data.Clear()
}

func (c *WeakCache[K, V]) Compute(
	key K,
	valueFn func(oldValue V, loaded bool) (newValue V, op ComputeOp),
) {
	c.data.Compute(
		key,
		func(oldValue cacheEntry[V], loaded bool) (newValue cacheEntry[V], op xsync.ComputeOp) {
			val := weak.Pointer[V](oldValue).Value()
			if val == nil {
				return oldValue, DeleteOp
			}

			newVal, op := valueFn(*val, loaded)
			if op != UpdateOp {
				return oldValue, op
			}

			return cacheEntry[V](weak.Make(&newVal)), op
		},
	)
}

func (c *WeakCache[K, V]) Range(f func(key K, value V) bool) {
	c.data.Range(
		func(key K, value cacheEntry[V]) bool {
			val := weak.Pointer[V](value).Value()
			if val == nil {
				return true
			}

			return f(key, *val)
		},
	)
}

func (c *WeakCache[K, V]) Store(key K, value *V) {
	runtime.AddCleanup(value, c.makeCleanup(key), nil)

	c.data.Store(key, cacheEntry[V](weak.Make(value)))
}

func (c *WeakCache[K, V]) makeCleanup(key K) func(*V) {
	return func(_ *V) {
		c.data.Compute(
			key,
			func(oldValue cacheEntry[V], loaded bool) (cacheEntry[V], ComputeOp) {
				op := xsync.CancelOp
				if weak.Pointer[V](oldValue).Value() == nil {
					op = xsync.DeleteOp
				}

				return oldValue, op
			},
		)
	}
}
