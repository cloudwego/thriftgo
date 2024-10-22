# Copyright 2024 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -e
thriftgo -g fastgo:no_default_serdes=true,gen_setter=true -o=. ./testdata.thrift
cd testdata
rm -f go.mod
go mod init thriftgo/test/fastgo/testdata
go mod tidy
go build -v ./...
