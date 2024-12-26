package pokecache

import (
	"testing"
	"time"
)

func TestCache_AddAndGet(t *testing.T) {
	// Initialize cache with a 1-second expiration
	cache := NewCache(1 * time.Second)

	// Add a cache entry
	cache.Add("pokemon:1", []byte("Pikachu"))

	// Retrieve the entry
	value, exists := cache.Get("pokemon:1")
	if !exists {
		t.Errorf("Expected cache to contain pokemon:1")
	}
	expected := "Pikachu"
	if string(value) != expected {
		t.Errorf("Expected %s, got %s", expected, value)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	// Initialize cache with a 1-second expiration
	cache := NewCache(1 * time.Second)

	// Try to get a non-existent key
	_, exists := cache.Get("pokemon:2")
	if exists {
		t.Errorf("Expected cache to not contain pokemon:2")
	}
}

func TestCache_Expiration(t *testing.T) {
	// Initialize cache with a 1-second expiration
	cache := NewCache(1 * time.Second)

	// Add a cache entry
	cache.Add("pokemon:1", []byte("Pikachu"))

	// Wait for 2 seconds to let the cache expire
	time.Sleep(2 * time.Second)

	// Attempt to retrieve the expired value
	_, exists := cache.Get("pokemon:1")
	if exists {
		t.Errorf("Expected pokemon:1 to be expired")
	}
}

func TestCache_ReapLoop(t *testing.T) {
	// Initialize cache with a 1-second expiration
	cache := NewCache(1 * time.Second)

	// Add some cache entries
	cache.Add("pokemon:1", []byte("Pikachu"))
	cache.Add("pokemon:2", []byte("Charmander"))

	// Wait for 2 seconds to ensure the entries expire
	time.Sleep(2 * time.Second)

	// Try to retrieve the expired entries
	_, exists := cache.Get("pokemon:1")
	if exists {
		t.Errorf("Expected pokemon:1 to be removed after expiration")
	}

	_, exists = cache.Get("pokemon:2")
	if exists {
		t.Errorf("Expected pokemon:2 to be removed after expiration")
	}
}

func TestCache_NoExpirationBeforeInterval(t *testing.T) {
	// Initialize cache with a 2-second expiration
	cache := NewCache(2 * time.Second)

	// Add a cache entry
	cache.Add("pokemon:1", []byte("Pikachu"))

	// Wait for only 1 second (less than the expiration interval)
	time.Sleep(1 * time.Second)

	// Ensure the entry still exists
	_, exists := cache.Get("pokemon:1")
	if !exists {
		t.Errorf("Expected pokemon:1 to still be in cache before expiration")
	}
}
