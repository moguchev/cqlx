// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package qb

import "testing"

func BenchmarkDeleteBuilder(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		Delete("cycling.cyclist_name").Columns("id", "user_uuid", "firstname", "stars").Where(Eq("id")).ToCql()
	}
}
