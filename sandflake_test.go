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
	"strconv"

	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	workerID := newWorkerID()
	var seq uint32 = maxSequence / 2
	date := time.Date(2017, 5, 27, 0, 0, 0, 20e6, time.UTC)

	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBytes := generateRandomBytes(source)
	d := NewID(date, workerID, seq, randomBytes)
	fmt.Println(d.String())

	out, err := Parse(d.String())
	require.Nil(t, err)
	require.True(t, date.Equal(out.Time()))
	require.Equal(t, seq, out.Sequence())
	require.Equal(t, workerID, out.WorkerID())
}

func TestClockSequences(t *testing.T) {
	origin := time.Date(2017, 5, 27, 0, 0, 0, 0, time.UTC)
	var tests = []struct {
		startTime            time.Time
		times                []time.Time
		expectedLastSequence int
	}{
		{
			times: []time.Time{
				origin,
				origin.Add(20 * time.Microsecond),
			},
			expectedLastSequence: 1,
		},
		{
			times: []time.Time{
				origin.Add(20 * time.Millisecond),
				origin,
			},
			expectedLastSequence: 0,
		},
		{
			times: []time.Time{
				origin,
				origin.Add(1 * time.Millisecond),
			},
			expectedLastSequence: 0,
		},
		{
			startTime:            origin,
			times:                makeTimes(origin, 20, time.Microsecond),
			expectedLastSequence: 20,
		},
		{
			startTime:            origin,
			times:                makeTimes(origin, 7, 250*time.Microsecond),
			expectedLastSequence: 3,
		},
	}

	for i, test := range tests {
		t.Run("test #"+strconv.Itoa(i), func(t *testing.T) {
			var id ID
			g := Generator{}
			if !test.startTime.IsZero() {
				g.lastTime = test.startTime
			}
			for _, t := range test.times {
				g.clock = mockClock(t)
				id = g.Next()
			}
			require.Equal(t, test.expectedLastSequence, int(id.Sequence()))
		})
	}
}

func TestOrdering(t *testing.T) {
	var g Generator
	d1 := g.Next()
	d2 := g.Next()

	require.Equal(t, 1, bytes.Compare(d2[:], d1[:]))
	require.True(t, d1.String() < d2.String()) // lexically sortable
}

func TestColision(t *testing.T) {
	var g Generator
	history := map[int64]map[uint32]ID{}

	for i := 0; i < 1e6; i++ {
		id := g.Next()
		ms := id.Time().UnixNano() / timeUnit
		seq := id.Sequence()
		if _, ok := history[ms]; !ok {
			history[ms] = make(map[uint32]ID)
		}

		last, ok := history[ms][seq]
		require.False(t, ok, "last: %v current: %v", last, id)
		history[ms][seq] = id
	}
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

func makeTimes(origin time.Time, n int, space time.Duration) []time.Time {
	times := make([]time.Time, n)
	for i := 0; i < n; i++ {
		times[i] = origin.Add(time.Duration(i+1) * space)
	}

	return times
}
