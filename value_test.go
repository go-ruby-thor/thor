// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import (
	"reflect"
	"testing"
)

func TestOrderedMap(t *testing.T) {
	m := NewOrderedMap()
	m.Set("a", "1")
	m.Set("b", "2")
	m.Set("a", "3") // overwrite keeps position
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Fatalf("keys=%v", m.Keys())
	}
	if v, ok := m.Get("a"); v != "3" || !ok {
		t.Fatalf("a=%v ok=%v", v, ok)
	}
	if _, ok := m.Get("z"); ok {
		t.Fatal("z should be absent")
	}
	if !m.Has("b") || m.Has("z") {
		t.Fatal("Has broken")
	}
	if m.Len() != 2 {
		t.Fatalf("len=%d", m.Len())
	}
}

func TestOrderedMapMerge(t *testing.T) {
	a := NewOrderedMap()
	a.Set("x", "1")
	b := NewOrderedMap()
	b.Set("y", "2")
	b.Set("x", "9")
	a.Merge(b)
	if !reflect.DeepEqual(a.Keys(), []string{"x", "y"}) {
		t.Fatalf("keys=%v", a.Keys())
	}
	if v, _ := a.Get("x"); v != "9" {
		t.Fatalf("x=%v", v)
	}
}

func TestValueMap(t *testing.T) {
	m := NewValueMap()
	m.Set("a", 1)
	m.Set("b", "two")
	m.Set("a", 3) // overwrite keeps position
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Fatalf("keys=%v", m.Keys())
	}
	if v, ok := m.Get("a"); v != 3 || !ok {
		t.Fatalf("a=%v ok=%v", v, ok)
	}
	if !m.Has("b") || m.Has("z") {
		t.Fatal("Has broken")
	}
	if m.Len() != 2 {
		t.Fatalf("len=%d", m.Len())
	}
}
