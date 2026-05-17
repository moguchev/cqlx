// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package qb

import "testing"

func BenchmarkBatchBuilder(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		Batch().Add(mockBuilder{"INSERT INTO cycling.cyclist_name (id,user_uuid,firstname) VALUES (?,?,?) ", []string{"id", "user_uuid", "firstname"}}).ToCql()
	}
}
