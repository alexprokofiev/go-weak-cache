package goweakcache

import (
	"time"
)

type expirableCacheEntry[V any] struct {
	ptr       *V
	expiredAt *time.Time
}

type ExpiringWeakCache[K comparable, V any] struct {
	weakCache *WeakCache[K, expirableCacheEntry[V]]
}

func NewWithExpiration[K comparable, V any]() *ExpiringWeakCache[K, V] {
	return &ExpiringWeakCache[K, V]{
		New[K, expirableCacheEntry[V]](),
	}
}

func (c *ExpiringWeakCache[K, V]) Load(key K) (*V, bool) {
	entry, ok := c.weakCache.Load(key)
	if !ok {
		return nil, false
	}

	if entry.expiredAt.IsZero() {
		return entry.ptr, true
	}

	if entry.expiredAt.Before(time.Now()) {
		return nil, false
	}

	return entry.ptr, true
}

func (c *ExpiringWeakCache[K, V]) Delete(key K) {
	c.weakCache.Delete(key)
}

func (c *ExpiringWeakCache[K, V]) Clear() {
	c.weakCache.Clear()
}

func (c *ExpiringWeakCache[K, V]) Compute(
	key K,
	valueFn func(oldValue V, loaded bool) (newValue V, expiredAt *time.Time, op ComputeOp),
) {
	c.weakCache.Compute(
		key,
		func(oldValue expirableCacheEntry[V], loaded bool) (newValue expirableCacheEntry[V], op ComputeOp) {
			if oldValue.expiredAt.IsZero() {
				newVal, expiredAt, op := valueFn(*oldValue.ptr, loaded)

				return expirableCacheEntry[V]{
						ptr:       &newVal,
						expiredAt: expiredAt,
					},
					op
			}

			if oldValue.expiredAt.Before(time.Now()) {
				return oldValue, DeleteOp
			}

			newVal, expiredAt, op := valueFn(*oldValue.ptr, loaded)

			return expirableCacheEntry[V]{
					ptr:       &newVal,
					expiredAt: expiredAt,
				},
				op
		},
	)
}

func (c *ExpiringWeakCache[K, V]) Range(key K, f func(key K, value V) bool) {
	c.weakCache.Range(
		func(key K, value expirableCacheEntry[V]) bool {
			if value.expiredAt.IsZero() {
				return f(key, *value.ptr)
			}

			if value.expiredAt.Before(time.Now()) {
				return true
			}

			return f(key, *value.ptr)
		},
	)
}

func (c *ExpiringWeakCache[K, V]) StoreEX(key K, value *V, expiredAt time.Time) {
	c.weakCache.Store(
		key,
		&expirableCacheEntry[V]{
			ptr:       value,
			expiredAt: &expiredAt,
		},
	)
}

func (c *ExpiringWeakCache[K, V]) Store(key K, value *V) {
	c.StoreEX(key, value, time.Time{})
}
