# Copyright 2023 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#!/bin/bash

set -e

basic_file_dir="test_cases"
basic_files=($(find "$basic_file_dir" -name "*.thrift" -type f -print))
basic_total=${#basic_files[@]}

rm -rf trimmer_test
mkdir trimmer_test

cd trimmer_test
go mod init trimmer

error_tests=()
test_result="test_result.txt"
rm -f $test_result
for i in "${!basic_files[@]}"; do
    test_num=$(($i+1))
    echo "Test [$test_num/$basic_total]:   ${basic_files[$i]}"
    rm -rf gen-go

    cd ..
    stdout_file=$(mktemp)
    stderr_file=$(mktemp)
    back="trimmer_test/back.thrift"
    cp "${basic_files[$i]}" "${back}-$i"
    if go run ../../. -g go -o trimmer_test/gen-go -r "${basic_files[$i]}"> "$(mktemp)" 2> "$(mktemp)"; then
        rm -rf trimmer_test/gen-go
        if go run . -o "${basic_files[$i]}" "${basic_files[$i]}"> "$stdout_file" 2> "$stderr_file"; then
            rm "$stdout_file" "$stderr_file"
            else
                error_tests+=("$test_num: ${basic_files[$i]}")
                echo "Test dump ${basic_files[$i]} failed! Error output:"
                cat "$stderr_file"
                echo "Test dump $i ${basic_files[$i]} failed! Error output:" >> "${test_result}"
                cat "$stderr_file" >> "${test_result}"
                echo "=======">> "${test_result}"
        fi
            if go run ../../. -g go -o trimmer_test/gen-go -r "${basic_files[$i]}" > "$stdout_file" 2> "$stderr_file"; then
                go mod edit -replace=github.com/apache/thrift=github.com/apache/thrift@v0.13.0
                go mod tidy
                go build ./...
                rm "$stdout_file" "$stderr_file"
            else
                error_tests+=("$test_num: ${basic_files[$i]}")
                echo "Test compile output of ${basic_files[$i]} failed! Error output:"
                cat "$stderr_file"
                echo "Test compile output of $i ${basic_files[$i]} failed! Error output:" >> "${test_result}"
                cat "$stderr_file" >> "${test_result}"
                echo "=======">> "${test_result}"
            fi
        else
          echo "thrift file incorrect, ignored.."
    fi
    cp "${basic_files[$i]}" "trimmer_test/$i-out.thrift"
    cp "${back}-$i" "${basic_files[$i]}"
    rm -f "${back}-$i"
    cd trimmer_test
done

if [ ${#error_tests[@]} -eq 0 ]; then
    echo "All tests passed!"
else
    echo "The following tests failed:"
    printf '%s\n' "${error_tests[@]}"
fi