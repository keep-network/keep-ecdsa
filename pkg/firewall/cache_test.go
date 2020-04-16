package firewall

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	cache := newTimeCache(time.Minute)

	cache.add("test")

	if !cache.has("test") {
		t.Fatal("should have 'test' key")
	}
}

func TestConcurrentAdd(t *testing.T) {
	cache := newTimeCache(time.Minute)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(item int) {
			cache.add(strconv.Itoa(item))
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		if !cache.has(strconv.Itoa(i)) {
			t.Fatalf("should have '%v' key", i)
		}
	}
}

func TestExpiration(t *testing.T) {
	cache := newTimeCache(500 * time.Millisecond)
	for i := 0; i < 6; i++ {
		cache.add(strconv.Itoa(i))
		time.Sleep(100 * time.Millisecond)
	}

	if cache.has(strconv.Itoa(0)) {
		t.Fatal("should have dropped '0' key from the cache already")
	}
}
