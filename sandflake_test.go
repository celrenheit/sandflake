/*
Copyright 2017 Salim Alami

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

package sandflake

import (
	"bytes"
	"fmt"
	"math/rand"

	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	workerID := newWorkerID()
	var seq uint32 = maxSequence / 2
	date := time.Date(2017, 5, 27, 0, 0, 0, 20e6, time.UTC)
	ts := uint64(date.UnixNano() / timeUnit)

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := NewID(ts, workerID, seq, source)
	fmt.Println(d.String())

	out, err := Parse(d.String())
	require.Nil(t, err)
	require.Equal(t, ts, uint64(out.Time().UnixNano()/timeUnit))
	require.Equal(t, seq, out.Sequence())
	require.Equal(t, workerID, out.WorkerID())
}

func TestOrdering(t *testing.T) {
	var g Generator
	d1 := g.Next()
	d2 := g.Next()

	require.Equal(t, 1, bytes.Compare(d2[:], d1[:]))
	require.True(t, d1.String() < d2.String()) // lexically sortable
}

func BenchmarkNext(b *testing.B) {
	b.ReportAllocs()
	var g Generator
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Next()
		}
	})
}

func BenchmarkToString(b *testing.B) {
	b.ReportAllocs()
	var g Generator
	id := g.Next()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var _ = id.String()
		}
	})
}

func BenchmarkParseID(b *testing.B) {
	var g Generator
	id := g.Next()
	idStr := id.String()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Parse(idStr)
		}
	})
}
