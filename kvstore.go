// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trr

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"
)

// a key-value store backed by raft
type kvstore struct {
	proposeC  chan<- string // channel for proposing updates
	mu        sync.RWMutex
	kvStore   map[string][]byte // current committed key-value pairs
	needFlush bool
}

type kv struct {
	Key string
	Val []byte
}

func newKVStore(proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *kvstore {
	s := &kvstore{proposeC: proposeC, kvStore: make(map[string][]byte)}
	// replay log into key-value map
	s.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go s.readCommits(commitC, errorC)
	return s
}

func (s *kvstore) Lookup(key string) ([]byte, bool) {
	//log.Println("Kv find K=", key, " map:", s.kvStore)
	flushLimit := 0
	for s.needFlush && flushLimit < 10 {
		log.Printf("** wait a little bit to make sure data flush **\n")
		time.Sleep(500 * time.Millisecond)
		flushLimit++
	}
	s.mu.RLock()
	v, ok := s.kvStore[key]
	s.mu.RUnlock()
	return v, ok
}

func (s *kvstore) Propose(k string, v []byte) {
	s.needFlush = true
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv{k, v}); err != nil {
		log.Fatal(err)
	}
	s.proposeC <- string(buf.Bytes())
}

func (s *kvstore) readCommits(commitC <-chan *string, errorC <-chan error) {

	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			return
		}

		var data_kv kv
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&data_kv); err != nil {
			log.Fatalf("trr: could not decode message (%v)", err)
		}
		s.mu.Lock()
		//log.Println("readCommits k->", data_kv.Key, " v->", data_kv.Val)
		s.kvStore[data_kv.Key] = data_kv.Val
		s.mu.Unlock()
		s.needFlush = false
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}

}
