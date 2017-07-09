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
	"io"
	"math/rand"
	"sync"
	"time"
)

type Generator struct {
	mu       sync.Mutex
	workerID WorkerID
	lastTime time.Time
	sequence uint32
	once     sync.Once
	reader   io.Reader
	clock    clock
}

// Next returns the next id.
// It returns an error if New() fails.
// It is safe for concurrent use.
func (g *Generator) Next() ID {
	g.once.Do(func() {
		g.workerID = newWorkerID()
		g.reader = rand.New(rand.NewSource(time.Now().UnixNano()))
		if g.clock == nil {
			g.clock = stdClock{}
		}
	})

	now := g.clock.Now().UTC()
	g.mu.Lock()

	if sub := now.Sub(g.lastTime); sub >= 0 && sub < time.Millisecond {
		g.sequence++
	} else {
		g.lastTime = now
		g.sequence = 0
	}

	if g.sequence > maxSequence {
		// reset sequence
		g.sequence = 0
	}

	wid := g.workerID
	seq := g.sequence
	g.mu.Unlock()

	return NewID(now, wid, seq, g.reader)
}

type clock interface {
	Now() time.Time
}

type stdClock struct{}

func (c stdClock) Now() time.Time { return time.Now() }

type mockClock time.Time

func (t mockClock) Now() time.Time { return time.Time(t) }
