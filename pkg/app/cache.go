package app

import (
	"errors"
	"fmt"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	store map[any]*entry
}

type entry struct {
	mu       sync.Mutex
	value    any
	lazyFunc func() (any, error)
	ready    bool
	err      error
}

var (
	ErrKeyExists   = errors.New("key already exists")
	ErrKeyNotFound = errors.New("key not found")
	ErrInitFailed  = errors.New("lazy initialization failed")
	ErrContextDone = errors.New("context canceled or timed out")
)

func NewCache() *Cache {
	return &Cache{
		store: make(map[any]*entry),
	}
}

func (c *Cache) Put(key, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.store[key]; exists {
		return ErrKeyExists
	}

	cacheEntry := &entry{}

	if fn, ok := value.(func() (any, error)); ok {
		cacheEntry.lazyFunc = fn
	} else {
		cacheEntry.value = value
		cacheEntry.ready = true
	}

	c.store[key] = cacheEntry

	return nil
}

func (c *Cache) Get(key any) (any, error) {
	c.mu.RLock()
	cacheEntry, ok := c.store[key]
	c.mu.RUnlock()

	if !ok {
		return nil, ErrKeyNotFound
	}

	cacheEntry.mu.Lock()
	defer cacheEntry.mu.Unlock()

	// Wait for readiness
	for !cacheEntry.ready {
		if cacheEntry.lazyFunc != nil {
			// This goroutine initializes
			val, err := cacheEntry.lazyFunc()
			if err != nil {
				cacheEntry.err = fmt.Errorf("%w: %w", ErrInitFailed, err)
			} else {
				cacheEntry.value = val
			}

			cacheEntry.ready = true
			cacheEntry.lazyFunc = nil

			break
		}
	}

	if cacheEntry.err != nil {
		return nil, cacheEntry.err
	}

	return cacheEntry.value, nil
}

func (c *Cache) Delete(key any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.store[key]; !exists {
		return ErrKeyNotFound
	}

	delete(c.store, key)

	return nil
}
