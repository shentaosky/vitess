/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInitialState(t *testing.T) {
	cache := NewLRUCache(5)
	l, c, _ := cache.Stats()
	if l != 0 {
		t.Errorf("length = %v, want 0", l)
	}
	if c != 5 {
		t.Errorf("capacity = %v, want 5", c)
	}
}

func TestSetInsertsValue(t *testing.T) {
	cache := NewLRUCache(100)
	data := "0"
	key := "key"
	cache.Set(key, data)

	v, ok := cache.Get(key)
	if !ok || v != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}

	k := cache.Keys()
	if len(k) != 1 || k[0] != key {
		t.Errorf("Cache.Keys() returned incorrect values: %v", k)
	}
	values := cache.Items()
	if len(values) != 1 || values[0].Key != key {
		t.Errorf("Cache.Values() returned incorrect values: %v", values)
	}
}

func TestSetIfAbsent(t *testing.T) {
	cache := NewLRUCache(100)
	data := "0"
	key := "key"
	cache.SetIfAbsent(key, data)

	v, ok := cache.Get(key)
	if !ok || v != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}

	cache.SetIfAbsent(key, "1")

	v, ok = cache.Get(key)
	if !ok || v != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}
}

func TestGetValueWithMultipleTypes(t *testing.T) {
	cache := NewLRUCache(100)
	data := "0"
	key := "key"
	cache.Set(key, data)

	v, ok := cache.Get("key")
	if !ok || v != data {
		t.Errorf("Cache has incorrect value for \"key\": %v != %v", data, v)
	}

	v, ok = cache.Get(string([]byte{'k', 'e', 'y'}))
	if !ok || v != data {
		t.Errorf("Cache has incorrect value for []byte {'k','e','y'}: %v != %v", data, v)
	}
}

func TestSetUpdatesSize(t *testing.T) {
	cache := NewLRUCache(100)
	emptyValue := ""
	key := "key1"
	cache.Set(key, emptyValue)
	if length, _, _ := cache.Stats(); length == 0 {
		t.Errorf("cache.Length() = %v, expected not 0", length)
	}
}

func TestSetWithOldKeyUpdatesValue(t *testing.T) {
	cache := NewLRUCache(100)
	firstValue := "0"
	key := "key1"
	cache.Set(key, firstValue)
	someValue := "1"
	cache.Set(key, someValue)

	v, ok := cache.Get(key)
	if !ok || v != someValue {
		t.Errorf("Cache has incorrect value: %v != %v", someValue, v)
	}
}

func TestGetNonExistent(t *testing.T) {
	cache := NewLRUCache(100)

	if _, ok := cache.Get("notthere"); ok {
		t.Error("Cache returned a notthere value after no inserts.")
	}
}

func TestPeek(t *testing.T) {
	cache := NewLRUCache(2)
	val1 := "1"
	cache.Set("key1", val1)
	val2 := "2"
	cache.Set("key2", val2)
	// Make key1 the most recent.
	cache.Get("key1")
	// Peek key2.
	if v, ok := cache.Peek("key2"); ok && v != val2 {
		t.Errorf("key2 received: %v, want %v", v, val2)
	}
	// Push key2 out
	cache.Set("key3", "3")
	if v, ok := cache.Peek("key2"); ok {
		t.Errorf("key2 received: %v, want absent", v)
	}
}

func TestDelete(t *testing.T) {
	cache := NewLRUCache(100)
	value := "1"
	key := "key"

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value)

	if !cache.Delete(key) {
		t.Error("Expected item to be in cache.")
	}

	if length, _, _ := cache.Stats(); length != 0 {
		t.Errorf("cache.Length() = %v, expected 0", length)
	}

	if _, ok := cache.Get(key); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestClear(t *testing.T) {
	cache := NewLRUCache(100)
	value := "1"
	key := "key"

	cache.Set(key, value)
	cache.Clear()

	if length, _, _ := cache.Stats(); length != 0 {
		t.Errorf("cache.Length() = %v, expected 0 after Clear()", length)
	}
}

func TestCapacityIsObeyed(t *testing.T) {
	size := int64(3)
	cache := NewLRUCache(100)
	cache.SetCapacity(size)
	value := "1"

	// Insert up to the cache's capacity.
	cache.Set("key1", value)
	cache.Set("key2", value)
	cache.Set("key3", value)
	if length, _, _ := cache.Stats(); length != size {
		t.Errorf("cache.Length() = %v, expected %v", length, size)
	}
	// Insert one more; something should be evicted to make room.
	cache.Set("key4", value)
	if length, _, _ := cache.Stats(); length != size {
		t.Errorf("cache.Length() = %v, expected %v", length, size)
	}

	// Check json stats
	data := cache.StatsJSON()
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Errorf("cache.StatsJSON() returned bad json data: %v %v", data, err)
	}

	// Check various other stats
	if l := cache.Length(); l != size {
		t.Errorf("cache.StatsJSON() returned bad length: %v", l)
	}
	if c := cache.Capacity(); c != size {
		t.Errorf("cache.StatsJSON() returned bad length: %v", c)
	}

	// checks StatsJSON on nil
	cache = nil
	if s := cache.StatsJSON(); s != "{}" {
		t.Errorf("cache.StatsJSON() on nil object returned %v", s)
	}
}

func TestLRUIsEvicted(t *testing.T) {
	size := int64(3)
	cache := NewLRUCache(size)

	cache.Set("key1", "1")
	cache.Set("key2", "2")
	cache.Set("key3", "3")
	// lru: [key3, key2, key1]

	// Look up the elements. This will rearrange the LRU ordering.
	cache.Get("key3")
	beforeKey2 := time.Now()
	cache.Get("key2")
	afterKey2 := time.Now()
	cache.Get("key1")
	// lru: [key1, key2, key3]

	cache.Set("key0", "0")
	// lru: [key0, key1, key2]

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key3"); ok {
		t.Error("Least recently used element was not evicted.")
	}

	// Check oldest
	if o := cache.Oldest(); o.Before(beforeKey2) || o.After(afterKey2) {
		t.Errorf("cache.Oldest returned an unexpected value: got %v, expected a value between %v and %v", o, beforeKey2, afterKey2)
	}
}
