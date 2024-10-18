/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"runtime"
	"sync"
)

var onceApacheReport sync.Once

const DISABLE_ENV = "KITEX_APACHE_CODEC_DISABLE_WARNING"

func WarningApache(structName string) {
	onceApacheReport.Do(func() {
		var path string
		if _, file, line, ok := runtime.Caller(1); ok {
			path = fmt.Sprintf("%s:%d \n", file, line)
		}
		fmt.Printf("[Kitex Apache Codec Warning] %s is using apache codec, please disable it. Path: %s\n", structName, path)
	})
}
