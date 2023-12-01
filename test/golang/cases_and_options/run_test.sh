#! /bin/bash -e
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


mod=$(grep -m1 '^module.*' go.mod | cut -d' ' -f2)
output=gen-go # change this if the '-out' parameter is supplied
errors=/tmp/thriftgo_test_cases_errors.txt
prefix=${mod}/$output
opts=package_prefix=$prefix
entry=a.thrift
ignore="lib_name_conflict"

features=(\
    naming_style=thriftgo \
    naming_style=golint \
    naming_style=apache \
    ignore_initialisms \
    json_enum_as_text \ 
    gen_setter \
    gen_db_tag \
    omitempty_for_optional=false \
    #use_type_alias=false \
    validate_set=false \
    value_type_in_container \
    scan_value_for_enum \
    reorder_fields \
    typed_enum_string \
    keep_unknown_fields \
    gen_deep_equal \
    reserve_comments \
    compatible_names \
    nil_safe \
    frugal_tag \
    unescape_double_quote \
    json_stringer \ 
)

run_cases() {
    # Each folder that contains a $entry IDL is treated as a test case.
    ls -d */ | grep -v $output | egrep -v "$ignore" | while read c; do
        c=${c%/}
        if [ -d $output ]; then
            rm -r $output
        fi
        idl="${c}/$entry"
        if ! [ -f "$idl" ]; then
            continue
        fi
        thriftgo -r -g go:$opts $idl >& /dev/null && \
            go mod tidy >& /dev/null && \
            go build $mod/gen-go/... >& /dev/null
        if [ $? -ne 0 ]; then
            printf "  case '$c': \e[1;31;mfailed\e[m\n" 
            echo "  thriftgo -r -g go:$opts $idl && go mod tidy && go build $mod/gen-go/..." >&2
        else
            printf "  case '$c': \e[1;36;mok\e[m\n" 
        fi
    done
}

run() {
    out=$(run_cases 2>$errors)
    if [ -n "$out" ]; then
        echo
        echo "$out"

        err=$(cat $errors)
        if [ -n "$err" ]; then
            echo "$err"
            #exit 1 # uncomment this to make the script stop at first error
            return 1
        fi
    else
        echo OK
    fi
}

printf "\e[1;34;m(default setting)\e[m"
run

# Test all cases on each option.
for ext in ${features[@]}; do
    printf  "\e[1;34;m[$ext]\e[m" 
    opts=package_prefix=$prefix,$ext
    run
    err_cnt=$((err_cnt + $?))
done

exit $err_cnt
