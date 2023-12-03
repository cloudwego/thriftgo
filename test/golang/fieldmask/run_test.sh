#! /bin/bash

# Copyright 2022 CloudWeGo Authors
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

generate () {
    out=gen-$1
    opt="go:package_prefix=example.com/test/$out"
    idl=a.thrift
    if [ -d $out ]; then
        rm -rf $out
    fi
    mkdir -p $out

    if [ "$1" = "new" ]; then
        opt="$opt,with_field_mask,with_reflection"
    fi
    echo "thriftgo -g $opt -o $out $idl"
    thriftgo -g "$opt" -o $out $idl
}

generate old
generate new
go mod tidy
go test -v ./...
