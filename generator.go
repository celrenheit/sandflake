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
	lastMS   uint64
	sequence uint32
	once     sync.Once
	reader   io.Reader
}

// Next returns the next id.
// It returns an error if New() fails.
// It is safe for concurrent use.
func (g *Generator) Next() ID {
	g.once.Do(func() {
		g.workerID = newWorkerID()
		g.reader = rand.New(rand.NewSource(time.Now().UnixNano()))
	})

	now := time.Now().UTC()
	nowMS := uint64(now.UnixNano() / 1e6)

	g.mu.Lock()
	if nowMS == g.lastMS {
		g.sequence++
	} else {
		g.lastMS = nowMS
		g.sequence = 0
	}

	if g.sequence > maxSequence {
		// reset sequence
		g.sequence = 0
	}

	wid := g.workerID
	seq := g.sequence
	g.mu.Unlock()

	return NewID(nowMS, wid, seq, g.reader)
}
