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

package compatible

import "github.com/apache/thrift/lib/go/thrift"

// WrapException wraps an error into a thrift.TException.
func WrapException(err error) thrift.TException {
	if err == nil {
		return nil
	}

	if ex, ok := err.(thrift.TException); ok {
		return ex
	}

	return thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION, err.Error())
}
