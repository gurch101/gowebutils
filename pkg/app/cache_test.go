package app_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gurch101/gowebutils/pkg/app"
)

func TestPutAndGet(t *testing.T) {
	c := app.NewCache()
	key := "test"

	err := c.Put(key, "value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, err := c.Get(key)
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}

	if v != "value" {
		t.Errorf("expected value 'value', got %v", v)
	}
}

func TestPutDuplicate(t *testing.T) {
	c := app.NewCache()
	key := "dup"

	_ = c.Put(key, 1)
	err := c.Put(key, 2)

	if !errors.Is(err, app.ErrKeyExists) {
		t.Errorf("expected ErrKeyExists, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	c := app.NewCache()
	key := "del"

	_ = c.Put(key, "val")

	err := c.Delete(key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = c.Get(key)
	if !errors.Is(err, app.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestLazyInit(t *testing.T) {
	c := app.NewCache()
	key := "lazy"
	called := false

	err := c.Put(key, func() (any, error) {
		called = true
		return "lazyval", nil
	})
	if err != nil {
		t.Fatalf("unexpected error on put: %v", err)
	}

	v, err := c.Get(key)
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}

	if v != "lazyval" {
		t.Errorf("expected lazyval, got %v", v)
	}

	if !called {
		t.Errorf("expected lazy function to be called")
	}
}

func TestLazyInitError(t *testing.T) {
	c := app.NewCache()
	key := "fail"
	errMsg := "fail init"

	_ = c.Put(key, func() (any, error) {
		//nolint:err113
		return nil, errors.New(errMsg)
	})

	_, err := c.Get(key)
	if !errors.Is(err, app.ErrInitFailed) {
		t.Errorf("expected ErrInitFailed, got %v", err)
	}
}

func TestConcurrentLazyInit(t *testing.T) {
	c := app.NewCache()
	key := "concurrent"

	callCount := 0
	init := func() (any, error) {
		callCount++

		time.Sleep(50 * time.Millisecond)

		return "ready", nil
	}

	_ = c.Put(key, init)

	wg := sync.WaitGroup{}
	results := make([]any, 10)
	errors := make([]error, 10)

	for i := range results {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			v, err := c.Get(key)
			results[i] = v
			errors[i] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("got error on goroutine %d: %v", i, err)
		}

		if results[i] != "ready" {
			t.Errorf("unexpected value on goroutine %d: %v", i, results[i])
		}
	}

	if callCount != 1 {
		t.Errorf("lazy function called %d times; want 1", callCount)
	}
}
