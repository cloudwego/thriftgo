// Copyright 2024 CloudWeGo Authors
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

package trim

import (
	"github.com/cloudwego/thriftgo/pkg/test"
	"testing"
)

func Test_extractIDLComposeConfigFromDir(t *testing.T) {
	testcases := []struct {
		desc      string
		dir       string
		targetAST string
		expect    func(t *testing.T, cfg *IDLComposeArguments)
	}{
		{
			desc:      "only trim_config.yaml",
			dir:       "../test_cases/config/trim_config",
			targetAST: "test.thrift",
			expect: func(t *testing.T, cfg *IDLComposeArguments) {
				preserve := true
				matchGoName := true
				expect := &IDLComposeArguments{
					IDLs: map[string]*IDLArguments{
						"test.thrift": {
							Trimmer: &YamlArguments{
								Methods: []string{
									"TestService.func1",
									"TestService.func3",
								},
								Preserve: &preserve,
								PreservedStructs: []string{
									"useless",
								},
								MatchGoName: &matchGoName,
							},
						},
					},
				}
				test.DeepEqual(t, cfg, expect)
			},
		},
		{
			desc:      "only idl_compose.yaml",
			dir:       "../test_cases/config/idl_compose",
			targetAST: "test1.thrift",
			expect: func(t *testing.T, cfg *IDLComposeArguments) {
				preserve := true
				matchGoNameTrue := true
				matchGoNameFalse := false
				expect := &IDLComposeArguments{
					IDLs: map[string]*IDLArguments{
						"test1.thrift": {
							Trimmer: &YamlArguments{
								Methods: []string{
									"TestService.func1",
									"TestService.func3",
								},
								Preserve: &preserve,
								PreservedStructs: []string{
									"useless",
								},
								MatchGoName: &matchGoNameTrue,
							},
						},
						"test2.thrift": {
							Trimmer: &YamlArguments{
								Preserve:    &preserve,
								MatchGoName: &matchGoNameFalse,
							},
						},
					},
				}
				test.DeepEqual(t, cfg, expect)
			},
		},
		{
			desc:      "both trim_config.yaml and idl_compose.yaml, only idl_compose.yaml will be used",
			dir:       "../test_cases/config/trim_config_and_idl_compose",
			targetAST: "test1.thrift",
			expect: func(t *testing.T, cfg *IDLComposeArguments) {
				preserve := true
				matchGoNameTrue := true
				matchGoNameFalse := false
				expect := &IDLComposeArguments{
					IDLs: map[string]*IDLArguments{
						"test1.thrift": {
							Trimmer: &YamlArguments{
								Methods: []string{
									"TestService.func1",
									"TestService.func3",
								},
								Preserve: &preserve,
								PreservedStructs: []string{
									"useless",
								},
								MatchGoName: &matchGoNameTrue,
							},
						},
						"test2.thrift": {
							Trimmer: &YamlArguments{
								Preserve:    &preserve,
								MatchGoName: &matchGoNameFalse,
							},
						},
					},
				}
				test.DeepEqual(t, cfg, expect)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			cfg := extractIDLComposeConfigFromDir(tc.dir, tc.targetAST)
			tc.expect(t, cfg)
		})
	}
}
