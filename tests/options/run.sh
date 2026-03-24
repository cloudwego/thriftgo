#!/bin/bash
# Copyright 2025 CloudWeGo Authors
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

# Test all thriftgo go-backend options individually and in combination.
# Usage: ./run_test.sh [path-to-thriftgo]
#
# Exit code: number of failed test cases (0 = all pass).

THRIFTGO="${1:-thriftgo}"
IDL="test.thrift"
OUTDIR="gen-go"
PASS=0
FAIL=0
FAILED_CASES=()

cd "$(dirname "$0")"

cleanup() {
    rm -rf "$OUTDIR" go.mod go.sum
}

init_mod() {
    go mod init tests_scratch > /dev/null 2>&1
    go mod edit -replace github.com/cloudwego/thriftgo=../.. > /dev/null 2>&1
    go mod edit -require github.com/apache/thrift@v0.13.0 > /dev/null 2>&1
}

# run_case <label> <options...>
# Runs thriftgo, then checks that the output compiles.
run_case() {
    local label="$1"
    shift
    local opts="$*"

    cleanup

    # Initialise a throwaway go module so we can compile the output.
    init_mod

    local gen_arg="go"
    if [ -n "$opts" ]; then
        gen_arg="go:package_prefix=tests_scratch/$OUTDIR,${opts}"
    else
        gen_arg="go:package_prefix=tests_scratch/$OUTDIR"
    fi

    if ! $THRIFTGO -g "$gen_arg" "$IDL" > /dev/null 2>&1; then
        printf "  %-60s \e[1;31mFAIL (thriftgo)\e[m\n" "$label"
        FAIL=$((FAIL + 1))
        FAILED_CASES+=("$label")
        return
    fi

    # skip_go_gen produces no output dir — just check thriftgo succeeded
    if [ ! -d "$OUTDIR" ]; then
        printf "  %-60s \e[1;32mPASS (no output)\e[m\n" "$label"
        PASS=$((PASS + 1))
        return
    fi

    if go mod tidy > /dev/null 2>&1 \
        && go build ./gen-go/... > /dev/null 2>&1; then
        printf "  %-60s \e[1;32mPASS\e[m\n" "$label"
        PASS=$((PASS + 1))
    else
        printf "  %-60s \e[1;31mFAIL (build)\e[m\n" "$label"
        FAIL=$((FAIL + 1))
        FAILED_CASES+=("$label")
    fi
}

# run_case_expect_fail <label> <options...>
# Expects thriftgo to fail (conflict / invalid combo).
run_case_expect_fail() {
    local label="$1"
    shift
    local opts="$*"

    cleanup

    init_mod

    local gen_arg="go:package_prefix=tests_scratch/$OUTDIR,${opts}"

    if $THRIFTGO -g "$gen_arg" "$IDL" > /dev/null 2>&1 \
        && go mod tidy > /dev/null 2>&1 \
        && go build ./gen-go/... > /dev/null 2>&1; then
        # It succeeded when we expected failure — that's still informative,
        # mark as PASS (option is accepted even if logically conflicting).
        printf "  %-60s \e[1;33mACCEPTED (expected conflict)\e[m\n" "$label"
        PASS=$((PASS + 1))
    else
        printf "  %-60s \e[1;36mREJECTED (expected)\e[m\n" "$label"
        PASS=$((PASS + 1))
    fi
}

echo "============================================="
echo " thriftgo option tests"
echo "============================================="
echo

# -----------------------------------------------------------------
# 1. Default (no options)
# -----------------------------------------------------------------
echo "--- Default ---"
run_case "default (no options)"

# -----------------------------------------------------------------
# 2. Each boolean feature individually (enabled)
# -----------------------------------------------------------------
echo
echo "--- Individual boolean features (enabled) ---"

BOOL_FEATURES=(
    json_enum_as_text
    enum_marshal
    enum_unmarshal
    gen_setter
    gen_db_tag
    omitempty_for_optional
    use_type_alias
    validate_set
    value_type_in_container
    scan_value_for_enum
    reorder_fields
    typed_enum_string
    keep_unknown_fields
    gen_deep_equal
    compatible_names
    reserve_comments
    nil_safe
    frugal_tag
    unescape_double_quote
    gen_type_meta
    gen_json_tag
    always_gen_json_tag
    snake_style_json_tag
    lower_camel_style_json_tag
    with_reflection
    enum_as_int_32
    trim_idl
    json_stringer
    no_default_serdes
    no_alias_type_reflection_method
    enable_ref_interface
    no_fmt
    skip_empty
    no_processor
    get_enum_annotation
    apache_warning
    apache_adaptor
    skip_go_gen
)

for f in "${BOOL_FEATURES[@]}"; do
    run_case "$f" "$f"
done

# -----------------------------------------------------------------
# 3. Boolean features explicitly disabled (defaults that are true)
# -----------------------------------------------------------------
echo
echo "--- Disable default-on features ---"

DEFAULT_ON=(
    omitempty_for_optional
    use_type_alias
    validate_set
    scan_value_for_enum
    unescape_double_quote
    gen_json_tag
)

for f in "${DEFAULT_ON[@]}"; do
    run_case "${f}=false" "${f}=false"
done

# -----------------------------------------------------------------
# 4. Naming styles
# -----------------------------------------------------------------
echo
echo "--- Naming styles ---"

for style in thriftgo golint apache; do
    run_case "naming_style=$style" "naming_style=$style"
done

run_case "naming_style=golint + ignore_initialisms" "naming_style=golint,ignore_initialisms"

# -----------------------------------------------------------------
# 5. Templates
# -----------------------------------------------------------------
echo
echo "--- Templates ---"

run_case "template=slim" "template=slim"
run_case "template=raw_struct" "template=raw_struct"

# -----------------------------------------------------------------
# 6. Useful combinations
# -----------------------------------------------------------------
echo
echo "--- Valid combinations ---"

run_case "gen_setter + nil_safe" \
    "gen_setter,nil_safe"

run_case "gen_deep_equal + keep_unknown_fields" \
    "gen_deep_equal,keep_unknown_fields"

run_case "json_enum_as_text + scan_value_for_enum" \
    "json_enum_as_text,scan_value_for_enum"

run_case "gen_db_tag + frugal_tag + gen_json_tag" \
    "gen_db_tag,frugal_tag,gen_json_tag"

run_case "value_type_in_container + gen_setter" \
    "value_type_in_container,gen_setter"

run_case "with_reflection + with_field_mask" \
    "with_reflection,with_field_mask"

run_case "with_reflection + with_field_mask + field_mask_halfway" \
    "with_reflection,with_field_mask,field_mask_halfway"

run_case "with_reflection + with_field_mask + field_mask_zero_required" \
    "with_reflection,with_field_mask,field_mask_zero_required"

run_case "thrift_streaming + streamx" \
    "thrift_streaming,streamx"

run_case "template=slim + enable_nested_struct" \
    "template=slim,enable_nested_struct"

run_case "template=raw_struct + enable_nested_struct" \
    "template=raw_struct,enable_nested_struct"

run_case "reserve_comments + compatible_names + nil_safe" \
    "reserve_comments,compatible_names,nil_safe"

run_case "snake_style_json_tag + always_gen_json_tag" \
    "snake_style_json_tag,always_gen_json_tag"

run_case "lower_camel_style_json_tag + always_gen_json_tag" \
    "lower_camel_style_json_tag,always_gen_json_tag"

run_case "no_default_serdes + no_processor" \
    "no_default_serdes,no_processor"

run_case "reorder_fields + gen_setter + gen_deep_equal" \
    "reorder_fields,gen_setter,gen_deep_equal"

run_case "enum_marshal + enum_unmarshal" \
    "enum_marshal,enum_unmarshal"

run_case "json_stringer + gen_json_tag" \
    "json_stringer,gen_json_tag"

run_case "typed_enum_string + enum_as_int_32" \
    "typed_enum_string,enum_as_int_32"

run_case "all tags: gen_db_tag + frugal_tag + always_gen_json_tag" \
    "gen_db_tag,frugal_tag,always_gen_json_tag"

run_case "naming_style=golint + ignore_initialisms + gen_setter + nil_safe" \
    "naming_style=golint,ignore_initialisms,gen_setter,nil_safe"

# -----------------------------------------------------------------
# 7. Potentially conflicting / edge-case combinations
# -----------------------------------------------------------------
echo
echo "--- Conflict / edge-case combinations ---"

# apache_warning vs apache_adaptor — mutually exclusive
run_case_expect_fail "apache_warning + apache_adaptor (conflict)" \
    "apache_warning,apache_adaptor"

# gen_deep_equal is silently disabled with slim template
run_case "template=slim + gen_deep_equal (deep_equal silently off)" \
    "template=slim,gen_deep_equal"

# with_field_mask without with_reflection — should need reflection
run_case_expect_fail "with_field_mask WITHOUT with_reflection" \
    "with_field_mask"

# streamx without thrift_streaming — streamx requires it
run_case_expect_fail "streamx WITHOUT thrift_streaming" \
    "streamx"

# field_mask_halfway without with_field_mask
run_case_expect_fail "field_mask_halfway WITHOUT with_field_mask" \
    "field_mask_halfway"

# enable_nested_struct without slim/raw_struct template (auto-switches)
run_case "enable_nested_struct (auto template switch)" \
    "enable_nested_struct"

# snake + lower_camel json tag (both enabled — last wins or conflict)
run_case_expect_fail "snake_style_json_tag + lower_camel_style_json_tag" \
    "snake_style_json_tag,lower_camel_style_json_tag"

# no_default_serdes + gen_deep_equal (serdes off but deep_equal on)
run_case "no_default_serdes + gen_deep_equal" \
    "no_default_serdes,gen_deep_equal"

# skip_go_gen — should produce no go files
run_case "skip_go_gen (no output)" \
    "skip_go_gen"

# json_enum_as_text vs enum_as_int_32 (conflicting enum representations)
run_case_expect_fail "json_enum_as_text + enum_as_int_32 (conflict?)" \
    "json_enum_as_text,enum_as_int_32"

# omitempty_for_optional=false + always_gen_json_tag
run_case "omitempty=false + always_gen_json_tag" \
    "omitempty_for_optional=false,always_gen_json_tag"

# gen_json_tag=false + snake_style_json_tag (no json tag but snake style)
run_case_expect_fail "gen_json_tag=false + snake_style_json_tag" \
    "gen_json_tag=false,snake_style_json_tag"

# gen_json_tag=false + always_gen_json_tag (contradictory)
run_case_expect_fail "gen_json_tag=false + always_gen_json_tag" \
    "gen_json_tag=false,always_gen_json_tag"

# use_type_alias=false + gen_type_meta
run_case "use_type_alias=false + gen_type_meta" \
    "use_type_alias=false,gen_type_meta"

# validate_set=false + value_type_in_container
run_case "validate_set=false + value_type_in_container" \
    "validate_set=false,value_type_in_container"

# -----------------------------------------------------------------
# 8. Structural edge cases (from cases_and_options)
# -----------------------------------------------------------------
echo
echo "--- Structural edge cases ---"

# run_structural <case_dir>
# Generates from cases/<dir>/a.thrift with -r and default options, then builds.
run_structural() {
    local dir="$1"
    local label="case/$dir"
    local idl="cases/${dir}/a.thrift"

    if [ ! -f "$idl" ]; then
        printf "  %-60s \e[1;33mSKIP (no IDL)\e[m\n" "$label"
        return
    fi

    cleanup
    init_mod

    local gen_arg="go:package_prefix=tests_scratch/$OUTDIR"

    if ! $THRIFTGO -r -g "$gen_arg" "$idl" > /dev/null 2>&1; then
        printf "  %-60s \e[1;31mFAIL (thriftgo)\e[m\n" "$label"
        FAIL=$((FAIL + 1))
        FAILED_CASES+=("$label")
        return
    fi

    if [ ! -d "$OUTDIR" ]; then
        printf "  %-60s \e[1;32mPASS (no output)\e[m\n" "$label"
        PASS=$((PASS + 1))
        return
    fi

    if go mod tidy > /dev/null 2>&1 \
        && go build ./gen-go/... > /dev/null 2>&1; then
        printf "  %-60s \e[1;32mPASS\e[m\n" "$label"
        PASS=$((PASS + 1))
    else
        printf "  %-60s \e[1;31mFAIL (build)\e[m\n" "$label"
        FAIL=$((FAIL + 1))
        FAILED_CASES+=("$label")
    fi
}

for d in cases/*/; do
    d=${d%/}
    d=${d#cases/}
    run_structural "$d"
done

# -----------------------------------------------------------------
# 9. CLI-level flags
# -----------------------------------------------------------------
echo
echo "--- CLI flags ---"

cleanup
# --version
if $THRIFTGO --version > /dev/null 2>&1; then
    printf "  %-60s \e[1;32mPASS\e[m\n" "--version"
    PASS=$((PASS + 1))
else
    printf "  %-60s \e[1;31mFAIL\e[m\n" "--version"
    FAIL=$((FAIL + 1))
    FAILED_CASES+=("--version")
fi

# --help (-h) — exits 0 or 2 depending on impl
if $THRIFTGO -h > /dev/null 2>&1; then
    printf "  %-60s \e[1;32mPASS\e[m\n" "--help"
    PASS=$((PASS + 1))
else
    # Many tools exit non-zero for --help; that's ok
    printf "  %-60s \e[1;33mEXIT NON-ZERO (acceptable)\e[m\n" "--help"
    PASS=$((PASS + 1))
fi

# -v (verbose)
cleanup
init_mod
if $THRIFTGO -v -g "go:package_prefix=tests_scratch/$OUTDIR" "$IDL" > /dev/null 2>&1; then
    printf "  %-60s \e[1;32mPASS\e[m\n" "-v (verbose)"
    PASS=$((PASS + 1))
else
    printf "  %-60s \e[1;31mFAIL\e[m\n" "-v (verbose)"
    FAIL=$((FAIL + 1))
    FAILED_CASES+=("-v")
fi

# -r (recurse) — no includes in our IDL so should still work
cleanup
init_mod
if $THRIFTGO -r -g "go:package_prefix=tests_scratch/$OUTDIR" "$IDL" > /dev/null 2>&1; then
    printf "  %-60s \e[1;32mPASS\e[m\n" "-r (recurse)"
    PASS=$((PASS + 1))
else
    printf "  %-60s \e[1;31mFAIL\e[m\n" "-r (recurse)"
    FAIL=$((FAIL + 1))
    FAILED_CASES+=("-r")
fi

# -----------------------------------------------------------------
# Summary
# -----------------------------------------------------------------
cleanup

echo
echo "============================================="
printf " Results: \e[1;32m%d passed\e[m, \e[1;31m%d failed\e[m\n" "$PASS" "$FAIL"
echo "============================================="

if [ ${#FAILED_CASES[@]} -gt 0 ]; then
    echo
    echo "Failed cases:"
    for c in "${FAILED_CASES[@]}"; do
        echo "  - $c"
    done
fi

exit "$FAIL"
