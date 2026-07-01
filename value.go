// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

// OrderedMap is an insertion-ordered string->string map, the Go shape of a
// Thor :hash option value. Thor builds hash values by walking "key:value"
// tokens left to right, so insertion order is meaningful and preserved here.
type OrderedMap struct {
	keys   []string
	values map[string]string
}

// NewOrderedMap returns an empty ordered map.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{values: map[string]string{}}
}

// Set inserts or replaces key. A repeated key keeps its original position and
// overwrites the value, matching Ruby Hash#[]= semantics (though Thor's parser
// rejects duplicate keys before they reach here).
func (m *OrderedMap) Set(key, val string) {
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = val
}

// Get returns the value for key and whether it was present.
func (m *OrderedMap) Get(key string) (string, bool) {
	v, ok := m.values[key]
	return v, ok
}

// Has reports whether key is present.
func (m *OrderedMap) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

// Len reports the number of entries.
func (m *OrderedMap) Len() int { return len(m.keys) }

// Keys returns the keys in insertion order. The slice must not be mutated.
func (m *OrderedMap) Keys() []string { return m.keys }

// Merge copies every entry of other into m in other's insertion order,
// mirroring how a repeatable :hash option accumulates via Hash#merge!.
func (m *OrderedMap) Merge(other *OrderedMap) {
	for _, k := range other.keys {
		m.Set(k, other.values[k])
	}
}

// Result is the outcome of a successful parse: the parsed option values keyed
// by option human name in declaration order, and the non-option remainder.
type Result struct {
	// Options maps each option's human name to its parsed value (a string,
	// int64, float64, bool, []string, or *OrderedMap), in declaration order.
	Options *ValueMap
	// Args holds the arguments left over after option parsing (Thor's #remaining).
	Args []string
}

// ValueMap is an insertion-ordered map from option human name to parsed value.
type ValueMap struct {
	keys   []string
	values map[string]any
}

// NewValueMap returns an empty value map.
func NewValueMap() *ValueMap {
	return &ValueMap{values: map[string]any{}}
}

// Set inserts or replaces key, preserving first-insertion position.
func (m *ValueMap) Set(key string, val any) {
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = val
}

// Get returns the value for key and whether it was present.
func (m *ValueMap) Get(key string) (any, bool) {
	v, ok := m.values[key]
	return v, ok
}

// Has reports whether key is present.
func (m *ValueMap) Has(key string) bool {
	_, ok := m.values[key]
	return ok
}

// Len reports the number of entries.
func (m *ValueMap) Len() int { return len(m.keys) }

// Keys returns the keys in insertion order. The slice must not be mutated.
func (m *ValueMap) Keys() []string { return m.keys }
