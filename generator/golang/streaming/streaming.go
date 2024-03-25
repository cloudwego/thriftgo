// Copyright 2023 CloudWeGo Authors
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

package streaming

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
)

// TODO this package should finally be migrated to Kitex tool together with server/client/processor templates.

const (
	// Annotation key used in Thrift IDL
	StreamingModeKey = "streaming.mode"

	// Streaming mode identifiers
	StreamingBidirectional = "bidirectional" // Bidirectional streaming API over HTTP2
	StreamingClientSide    = "client"        // Client-side streaming API over HTTP2
	StreamingServerSide    = "server"        // Server-side streaming API over HTTP2
	StreamingUnary         = "unary"         // Unary API over HTTP2, different from Kitex Thrift/Protobuf
)

// Streaming represents the streaming mode of a function
type Streaming struct {
	Mode                   string `thrift:"Mode,1" json:"Mode"`
	ClientStreaming        bool   `thrift:"ClientStreaming,2" json:"ClientStreaming"`
	ServerStreaming        bool   `thrift:"ServerStreaming,3" json:"ServerStreaming"`
	BidirectionalStreaming bool   `thrift:"BidirectionalStreaming,4" json:"BidirectionalStreaming"`
	Unary                  bool   `thrift:"Unary,5" json:"Unary"`
	IsStreaming            bool   `thrift:"IsStreaming,6" json:"IsStreaming"`
}

// ParseStreaming parses the streaming mode from a Thrift function parsed from IDL
func ParseStreaming(f *parser.Function) (s *Streaming, err error) {
	s = &Streaming{}
	for _, anno := range f.Annotations {
		if anno.Key != StreamingModeKey {
			continue
		}
		if len(anno.Values) != 1 {
			return nil, fmt.Errorf("%s has multiple values for annotation %v (at most 1 allowed)",
				f.Name, StreamingModeKey)
		}
		for _, value := range anno.Values {
			s.IsStreaming = true
			switch value {
			case StreamingClientSide:
				s.ClientStreaming = true
			case StreamingServerSide:
				s.ServerStreaming = true
			case StreamingBidirectional:
				s.ClientStreaming = true
				s.ServerStreaming = true
				s.BidirectionalStreaming = true
			case StreamingUnary:
				s.Unary = true
			default: // other types are not recognized
				return nil, fmt.Errorf("invalid value (%s) for annotation %v", value, StreamingModeKey)
			}
		}
		if s.IsStreaming && len(f.Arguments) != 1 {
			return nil, fmt.Errorf("streaming function %s should have exactly 1 argument", f.Name)
		}

		if s.BidirectionalStreaming {
			s.Mode = StreamingBidirectional
		} else if s.ServerStreaming {
			s.Mode = StreamingServerSide
		} else if s.ClientStreaming {
			s.Mode = StreamingClientSide
		} else if s.Unary {
			s.Mode = StreamingUnary
		}
	}
	return s, nil
}
