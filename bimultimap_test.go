package bimultimap

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBiMultiMap(t *testing.T) {
	sut := New[string, string]()
	expected := &BiMultiMap[string, string]{
		forward: make(map[string][]string),
		inverse: make(map[string][]string),
	}
	assert.Equal(t, expected, sut, "a new BiMultiMap should be empty")
}

func TestBiMultiMapPut(t *testing.T) {
	sut := New[string, string]()
	sut.Add("key", "value")

	assert.True(t, sut.KeyExists("key"), "the key should exist")
	assert.ElementsMatch(t, []string{"value"}, sut.LookupKey("key"), "the value associated with the key should be the correct one")

	assert.True(t, sut.ValueExists("value"), "the value should exist")
	assert.ElementsMatch(t, []string{"key"}, sut.LookupValue("value"), "the key associated with the value should be the correct one")
}

func TestBiMultiMapPutDup(t *testing.T) {
	sut := New[string, string]()
	sut.Add("key", "value")
	sut.Add("key", "value")

	assert.True(t, sut.KeyExists("key"), "the key should exist")
	assert.ElementsMatch(t, []string{"value"}, sut.LookupKey("key"), "the value associated with the key should not be duplicated")

	assert.True(t, sut.ValueExists("value"), "the value should exist")
	assert.ElementsMatch(t, []string{"key"}, sut.LookupValue("value"), "the key associated with the value should not be duplicated")
}

func TestBiMultiMapMultiPut(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()

	assert.True(t, sut.KeyExists("key1"), "key1 should exist")
	assert.True(t, sut.KeyExists("key2"), "key2 should exist")
	assert.ElementsMatch(t, []string{"value1", "value2"}, sut.LookupKey("key1"), "the values associated with the key should be the correct one")

	assert.True(t, sut.ValueExists("value1"), "value1 should exist")
	assert.True(t, sut.ValueExists("value2"), "value2 should exist")
	assert.ElementsMatch(t, []string{"key1", "key2"}, sut.LookupValue("value1"), "the keys associated with the value should be the correct one")
}

func TestBiMultiMapGetEmpty(t *testing.T) {
	sut := New[string, string]()
	assert.ElementsMatch(t, []string{}, sut.LookupValue("foo"), "a nonexistent key should return an empty slice")
	assert.ElementsMatch(t, []string{}, sut.LookupKey("foo"), "a nonexistent value should return an empty slice")

	sut.Add("key", "value")
	assert.ElementsMatch(t, []string{}, sut.LookupValue("foo"), "a nonexistent key should return an empty slice")
	assert.ElementsMatch(t, []string{}, sut.LookupKey("foo"), "a nonexistent value should return an empty slice")
}

func TestBiMultiMapDeleteKey(t *testing.T) {
	sut := New[string, string]()
	sut.Add("key", "value")

	value := sut.DeleteKey("key")

	assert.ElementsMatch(t, []string{}, sut.LookupKey("key"), "deleting a key should delete it")
	assert.ElementsMatch(t, []string{}, sut.LookupValue("value"), "a deleted key should delete its values")
	assert.ElementsMatch(t, []string{"value"}, value, "deleting a key should return its associated values")
}

func TestBiMultiMapDeleteValue(t *testing.T) {
	sut := New[string, string]()
	sut.Add("key", "value")

	value := sut.DeleteValue("value")

	assert.ElementsMatch(t, []string{}, sut.LookupKey("key"), "deleting a value should delete it")
	assert.ElementsMatch(t, []string{}, sut.LookupValue("value"), "a deleted value should delete its keys")
	assert.ElementsMatch(t, []string{"key"}, value, "deleting a value should return its associated keys")
}

func TestBiMultiMapDeleteMultiKey(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()

	sut.DeleteKey("key1")

	assert.ElementsMatch(t, []string{}, sut.LookupKey("key1"), "deleting a key should delete all its values")
	assert.ElementsMatch(t, []string{"value1", "value2"}, sut.LookupKey("key2"), "deleting a key should only delete its own values")
	assert.ElementsMatch(t, []string{"key2"}, sut.LookupValue("value1"), "deleting a key should delete the inverse")
}

func TestBiMultiMapDeleteMultiValue(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()

	sut.DeleteValue("value1")

	assert.ElementsMatch(t, []string{}, sut.LookupValue("value1"), "deleting a value should delete all its keys")
	assert.ElementsMatch(t, []string{"key1", "key2"}, sut.LookupValue("value2"), "deleting a value should only delete its own keys")
	assert.ElementsMatch(t, []string{"value2"}, sut.LookupKey("key1"), "deleting a value should delete the inverse")
}

func TestMultiMapDeleteKeyValue(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()

	sut.DeleteKeyValue("key1", "value1")

	assert.ElementsMatch(t, []string{"value2"}, sut.LookupKey("key1"), "deleting a key/value pair should delete only that key/value pair from its key")
	assert.ElementsMatch(t, []string{"value1", "value2"}, sut.LookupKey("key2"), "deleting a key/value pair should not affect other keys")
	assert.ElementsMatch(t, []string{"key2"}, sut.LookupValue("value1"), "deleting a key/value pair delete only that key/value pair from its value")
	assert.ElementsMatch(t, []string{"key1", "key2"}, sut.LookupValue("value2"), "deleting a key/value pair should not affect other values")
}

func TestBiMultiMapKeysValues(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()
	sut.Add("key3", "value3")

	keys := slices.Collect(sut.Keys())
	values := slices.Collect(sut.Values())

	assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys, "Keys() should return a slice containing the keys")
	assert.ElementsMatch(t, []string{"value1", "value2", "value3"}, values, "Values() should return a slice containing the keys")
}

func TestBiMultiMapClear(t *testing.T) {
	sut := biMultiMapWithMultipleKeysValues()
	sut.Clear()

	assert.Equal(t, len(slices.Collect(sut.Keys())), 0)
	assert.Equal(t, len(slices.Collect(sut.Values())), 0)
}

func TestBiMultiMapMerge(t *testing.T) {
	map1 := biMultiMapWithMultipleKeysValues()

	map2 := New[string, string]()
	map2.Add("key1", "value3")
	map2.Add("key3", "value3")
	map2.Add("key1", "value1")
	map2.Add("key3", "value1")
	map2.Add("key4", "value4")

	sut := map1.Merge(map2)

	assert.ElementsMatch(t, []string{"key1", "key2", "key3", "key4"}, slices.Collect(sut.Keys()))
	assert.ElementsMatch(t, []string{"value1", "value2", "value3"}, sut.LookupKey("key1"))
	assert.ElementsMatch(t, []string{"value1", "value3"}, sut.LookupKey("key3"))
	assert.ElementsMatch(t, []string{"value4"}, sut.LookupKey("key4"))

	assert.ElementsMatch(t, []string{"value1", "value2", "value3", "value4"}, slices.Collect(sut.Values()))
	assert.ElementsMatch(t, []string{"key1", "key2", "key3"}, sut.LookupValue("value1"))
	assert.ElementsMatch(t, []string{"key1", "key3"}, sut.LookupValue("value3"))
	assert.ElementsMatch(t, []string{"key4"}, sut.LookupValue("value4"))
}

func biMultiMapWithMultipleKeysValues() *BiMultiMap[string, string] {
	m := New[string, string]()
	m.Add("key1", "value1")
	m.Add("key1", "value2")
	m.Add("key2", "value1")
	m.Add("key2", "value2")

	return m
}
