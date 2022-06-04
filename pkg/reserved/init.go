// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reserved

import "sync"

var (
	lock sync.RWMutex
	all  = make(map[string][]string) // word => langs
)

// Register adds reserved words to the global set for the given language.
func Register(lang string, words ...string) {
	lock.Lock()
	defer lock.Unlock()
next:
	for _, w := range words {
		for _, l := range all[w] {
			if l == lang {
				continue next
			}
		}
		all[w] = append(all[w], lang)
	}
}

// Hit returns a list of the languages that keep the given word as reserved.
func Hit(word string) (res []string) {
	lock.RLock()
	defer lock.RUnlock()
	return all[word] // XXX: make a copy to avoid modification?
}
