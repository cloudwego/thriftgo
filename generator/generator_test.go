// Copyright 2021 CloudWeGo Authors
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

package generator

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/thriftgo/pkg/test"
)

type postProcessFunc func(path string, content []byte) ([]byte, error)

func (f postProcessFunc) PostProcess(path string, content []byte) ([]byte, error) {
	return f(path, content)
}

func TestAsyncPostProcess(t *testing.T) {
	// Test initialization
	p := newAsyncPostProcess(postProcessFunc(func(path string, content []byte) ([]byte, error) {
		time.Sleep(time.Millisecond)
		return append(content, "-processed"...), nil
	}))
	p.concurrency = 2

	// Test Add and OnFinished with success
	paths := []string{"file1.go", "file2.go", "file3.go", "file4.go"}
	contents := []string{"content1", "content2", "content3", "content4"}

	// Add files to process
	for i := range paths {
		p.Add(paths[i], contents[i])
	}

	// Track processed files
	var mu sync.Mutex
	processed := make(map[string]string)

	// Process files with OnFinished
	err := p.OnFinished(func(path string, content []byte) error {
		time.Sleep(time.Millisecond) // simulate processing
		mu.Lock()
		defer mu.Unlock()
		processed[path] = string(content)
		return nil
	})

	// Verify success
	test.Assert(t, err == nil)
	mu.Lock()
	test.Assert(t, len(processed) == len(paths), len(processed))
	mu.Unlock()
	for i, path := range paths {
		test.Assert(t, processed[path] == contents[i]+"-processed", processed[path])
	}

	// Test OnFinished with error
	p = newAsyncPostProcess(postProcessFunc(func(path string, content []byte) ([]byte, error) {
		return content, nil
	}))
	p.concurrency = 2
	p.Add("error.go", "content")
	p.Add("ok1.go", "content")
	p.Add("ok2.go", "content")
	p.Add("ok3.go", "content")
	p.Add("ok4.go", "content")

	err = p.OnFinished(func(path string, content []byte) error {
		time.Sleep(time.Millisecond)
		if path == "error.go" {
			return errors.New("test error")
		}
		return nil
	})

	// Verify error handling
	test.Assert(t, err != nil)
	test.Assert(t, err.Error() == "test error")
}
