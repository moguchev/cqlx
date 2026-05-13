// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package qb

import (
	"bytes"
	"testing"
)

func BenchmarkCmp(b *testing.B) {
	buf := bytes.Buffer{}
	b.ResetTimer()
	for b.Loop() {
		buf.Reset()
		c := cmps{
			Eq("id"),
			Lt("user_uuid"),
			LtOrEq("firstname"),
			Gt("stars"),
		}
		c.writeCql(&buf)
	}
}
