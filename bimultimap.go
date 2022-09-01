// Package bimultimap is a thread-safe bidirectional MultiMap. It associates a key with multiple
// values and each value with its corresponding keys.
package bimultimap

import (
	"sync"
)

// BiMultiMap is a thread-safe bidirectional multimap where neither the keys nor the values need to be unique
type BiMultiMap[K comparable, V comparable] struct {
	forward map[K][]V
	inverse map[V][]K
	mutex   sync.RWMutex
}

// New creates a new, empty biMultiMap
func New[K comparable, V comparable]() *BiMultiMap[K, V] {
	return &BiMultiMap[K, V]{
		forward: make(map[K][]V),
		inverse: make(map[V][]K),
	}
}

// LookupKey gets the values associated with a key, or an empty slice if the key does not exist
func (m *BiMultiMap[K, V]) LookupKey(key K) []V {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	values, found := m.forward[key]
	if !found {
		return make([]V, 0)
	}
	return values
}

// LookupValue gets the keys associated with a value, or an empty slice if the value does not exist
func (m *BiMultiMap[K, V]) LookupValue(value V) []K {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys, found := m.inverse[value]

	if !found {
		return make([]K, 0)
	}
	return keys
}

// Add adds a key/value pair
func (m *BiMultiMap[K, V]) Add(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	values, found := m.forward[key]
	if !found {
		values = make([]V, 0, 1)
	}

	// Value already exists for that key - early exit
	for _, v := range values {
		if v == value {
			return
		}
	}

	values = append(values, value)
	m.forward[key] = values

	keys, found := m.inverse[value]
	if !found {
		keys = make([]K, 0, 1)
	}
	keys = append(keys, key)
	m.inverse[value] = keys
}

// KeyExists returns true if a key exists in the map
func (m *BiMultiMap[K, V]) KeyExists(key K) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, found := m.forward[key]
	return found
}

// ValueExists returns true if a value exists in the map
func (m *BiMultiMap[K, V]) ValueExists(value V) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, found := m.inverse[value]
	return found
}

// DeleteKey deletes a key from the map and returns its associated values
func (m *BiMultiMap[K, V]) DeleteKey(key K) []V {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	values, found := m.forward[key]
	if !found {
		return make([]V, 0)
	}

	delete(m.forward, key)

	for _, v := range values {
		newKeys := deleteElement(m.inverse[v], key)
		m.inverse[v] = newKeys
	}

	return values
}

// DeleteValue deletes a value from the map and returns its associated keys
func (m *BiMultiMap[K, V]) DeleteValue(value V) []K {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys, found := m.inverse[value]
	if !found {
		return make([]K, 0)
	}

	delete(m.inverse, value)

	for _, k := range keys {
		newVals := deleteElement(m.forward[k], value)
		m.forward[k] = newVals
	}

	return keys
}

// DeleteKeyValue deletes a single key/value pair
func (m *BiMultiMap[K, V]) DeleteKeyValue(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, foundValue := m.forward[key]
	_, foundKey := m.inverse[value]

	if foundKey && foundValue {
		newVals := deleteElement(m.forward[key], value)
		if len(newVals) > 0 {
			m.forward[key] = newVals
		} else {
			delete(m.forward, key)
		}

		newKeys := deleteElement(m.inverse[value], key)
		if len(newKeys) > 0 {
			m.inverse[value] = newKeys
		} else {
			delete(m.inverse, value)
		}
	}
}

// Merge merges two BiMultiMap[K, V]s: returns a new BiMultiMap consisting of all the key/value pairs in
// this one and all key/value pairs in the other one
func (m *BiMultiMap[K, V]) Merge(other *BiMultiMap[K, V]) *BiMultiMap[K, V] {
	m.mutex.RLock()
	other.mutex.RLock()
	defer func() {
		other.mutex.RUnlock()
		m.mutex.RUnlock()
	}()

	res := New[K, V]()

	for _, k := range m.Keys() {
		for _, v := range m.LookupKey(k) {
			res.Add(k, v)
		}
	}

	for _, k := range other.Keys() {
		for _, v := range other.LookupKey(k) {
			res.Add(k, v)
		}
	}

	return res
}

// Clear clears all entries in the BiMultiMap[K, V]
func (m *BiMultiMap[K, V]) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.forward = make(map[K][]V)
	m.inverse = make(map[V][]K)
}

// Keys returns an unordered slice containing all of the map's keys
func (m *BiMultiMap[K, V]) Keys() []K {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys := make([]K, 0, len(m.forward))
	for k := range m.forward {
		keys = append(keys, k)
	}
	return keys
}

// Values returns an unordered slice containing all of the map's values
func (m *BiMultiMap[K, V]) Values() []V {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	values := make([]V, 0, len(m.inverse))
	for v := range m.inverse {
		values = append(values, v)
	}
	return values
}

// Helper function: delete an element from a slice if it exists
func deleteElement[T comparable](slice []T, element T) []T {
	newSlice := make([]T, 0, len(slice)-1)

	for _, val := range slice {
		if val != element {
			newSlice = append(newSlice, val)
		}
	}

	return newSlice
}
