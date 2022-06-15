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

package meta

import (
	"bytes"
	"context"
)

var _ RichTransport = (*MemoryTransport)(nil)

// MemoryTransport is a transport that manages data in memory.
// The zero value for MemoryTransport is ready to use.
type MemoryTransport struct {
	bytes.Buffer
}

// Close .
func (p *MemoryTransport) Close() error {
	p.Reset()
	return nil
}

// Flush .
func (p *MemoryTransport) Flush(ctx context.Context) error {
	return nil
}

// Open .
func (p *MemoryTransport) Open() error {
	return nil
}

// IsOpen .
func (p *MemoryTransport) IsOpen() bool {
	return true
}
