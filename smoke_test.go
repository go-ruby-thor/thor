// Copyright (c) the go-ruby-thor/thor authors
//
// SPDX-License-Identifier: BSD-3-Clause

package thor

import "testing"

func TestSmokeParse(t *testing.T) {
	name, err := NewOption("name", Option{Typ: String})
	if err != nil {
		t.Fatal(err)
	}
	opts := NewOptions([]*Option{name}, nil, false, false, Relations{})
	res, err := opts.Parse([]string{"--name", "David", "extra"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := res.Options.Get("name")
	if v != "David" {
		t.Fatalf("got %v", v)
	}
	if len(res.Args) != 1 || res.Args[0] != "extra" {
		t.Fatalf("args %v", res.Args)
	}
}
