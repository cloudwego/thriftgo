# trimmer

**trimmer** is a standalone tool that removes unused definitions from Thrift IDL files. It parses an IDL, keeps only the definitions reachable from service method signatures, and writes the result to a new `.thrift` file.

## What it does

- Walks a Thrift IDL and marks every struct, union, exception, enum, and typedef that is directly or transitively referenced by at least one service method (argument types, return types, exception types).
- Removes everything else and dumps the cleaned IDL to a file.
- Use it to reduce IDL size before code generation when a large shared IDL contains many definitions that a specific service does not use.

## Installation

**Prerequisites:** Go 1.18 or later. Ensure `$GOPATH/bin` (default: `~/go/bin`) is in your `PATH`.

```sh
go install github.com/cloudwego/thriftgo/tool/trimmer@latest
```

**Build from source:**

```sh
git clone https://github.com/cloudwego/thriftgo.git
cd thriftgo/tool/trimmer
go install
```

**Verify:**

```sh
trimmer --version
```

## Quick start

```sh
trimmer service.thrift
```

Produces `service_trimmed.thrift` in the same directory and prints a summary:

```
removed 42 unused structures with 187 fields
success, dump to service_trimmed.thrift
```

## Usage

```
trimmer [options] <file.thrift>
```

### Flags

| Flag | Short | Type | Default | Description | Example |
|---|---|---|---|---|---|
| `--version` | | bool | false | Print the version and exit. | `trimmer --version` |
| `--help` | `-h` | bool | false | Print help and exit. | `trimmer -h` |
| `--out` | `-o` | string | `<input>_trimmed.thrift` | Output file path. If a path to an existing directory is given, the input filename is used inside it. With `-r`, this must be a directory path; defaults to `./trimmed_idl` if omitted. | `trimmer -o ./out/service.thrift service.thrift` |
| `--recurse` | `-r` | string | | Base directory used to compute relative output paths when dumping included IDLs. All transitive includes are also dumped, preserving their relative layout under `-o`. The `-o` directory must be outside this path. | `trimmer -r ./idl -o ./trimmed_idl service.thrift` |
| `--method` | `-m` | string | | Keep only the specified method and its type dependencies. Form: `ServiceName.MethodName` (supports regexp2 patterns). Repeatable. If no service prefix is given, defaults to the only service (single-service IDL) or the last service (multi-service IDL). | `trimmer -m UserService.GetUser service.thrift` |
| `--preserve` | `-p` | bool | `true` | When `false`, disables all preservation mechanisms: `@preserve` comments, `preserved_structs`, and `preserved_files` in the config are all ignored. | `trimmer -p false service.thrift` |

### What is always kept

The following are never removed regardless of reachability or flags:

- **Constants** — all constants are kept unconditionally.
- **Typedefs** — all typedefs are kept unconditionally.
- **Enums** — all enums are kept unconditionally.
- **Included files that contain constants, enums, or typedefs** — an `include` statement is retained (and its IDL recursively trimmed) even if no struct types from it are directly referenced.
- **`@preserve`-annotated structs** — any struct, union, or exception whose reserved comments contain a line matching `// @preserve` or `# @preserve` (case-insensitive) is kept, unless `-p false` is passed.

### `-m` flag details

Without `-m`, every service method and its dependencies are kept. With one or more `-m` flags, only the listed methods are kept in the output; the service block itself remains but contains only those methods.

`-m` values are regexp2 patterns matched against `ServiceName.MethodName`. Multiple methods can be specified:

```sh
trimmer -m UserService.GetUser -m UserService.DeleteUser service.thrift
```

When no `ServiceName.` prefix is given:
- Single-service IDL: the only service name is prepended automatically.
- Multi-service IDL: the **last** service name in the file is used.

```sh
# IDL has one service — safe to omit prefix
trimmer -m GetUser -m CreateUser service.thrift
```

A warning is printed for any `-m` pattern that does not match any method in the IDL:

```
warning: method UserService.NoSuchMethod not found in service.thrift!
```

### `-r` flag details

`-r` enables recursive mode. The input file is still the positional argument. After trimming, trimmer walks all transitive includes and dumps each one as well.

The `-r` value is the **base directory** used to compute relative output paths. Each included IDL is written under `-o` at the same relative position it holds under `-r`. The `-o` directory must be outside the `-r` directory to avoid overwriting the source files.

If `-o` is not specified when `-r` is used, it defaults to `./trimmed_idl`.

```sh
# idl/service.thrift includes idl/base/base.thrift
# Output: trimmed_idl/service.thrift, trimmed_idl/base/base.thrift
trimmer -r ./idl -o ./trimmed_idl ./idl/service.thrift
```

## Input / Output behavior

- **Input:** A single `.thrift` IDL file as a positional argument.
- **Stdin:** Not used.
- **Stdout:** Progress messages and the removal summary (`removed N unused structures with Y fields` followed by `success, dump to <path>`). Note: the "structures" count includes service blocks and include statements, not only struct definitions.
- **Stderr:** Not used for normal output.
- **Output files:** One trimmed `.thrift` file per IDL processed (more when `-r` is used).

## Configuration

trimmer reads `trim_config.yaml` from the working directory if it exists. CLI flags take precedence over config file values.

**`trim_config.yaml` fields:**

| Field | Type | Default | Description |
|---|---|---|---|
| `methods` | `[]string` | | Same as `-m`. Applied only if no `-m` flags are passed. |
| `preserve` | `bool` | `true` | Same as `-p`. When `false`, disables all preservation: `@preserve` comments, `preserved_structs`, and `preserved_files` are all ignored. |
| `preserved_structs` | `[]string` | | Struct names to keep unconditionally. Disabled when `preserve: false`. When `match_go_name: true`, names are matched against the Go-converted name (`snake_case` → `PascalCase`). |
| `disable_preserve_comment` | `bool` | `false` | Skip scanning comments for `@preserve`. Improves performance on very large IDLs when comment-based preservation is not needed. |
| `match_go_name` | `bool` | `false` | Match `-m` method names and `preserved_structs` entries against Go-converted names instead of raw IDL names. |
| `preserved_files` | `[]string` | | Paths to included `.thrift` files whose structs should all be kept unconditionally. Disabled when `preserve: false`. |

**Example `trim_config.yaml`:**

```yaml
methods:
  - UserService.GetUser
  - UserService.CreateUser
preserve: true
preserved_structs:
  - CommonError
  - PageInfo
preserved_files:
  - idl/base/base.thrift
```

## Examples

**Trim a single file, output to a specific path:**

```sh
trimmer -o ./output/service_slim.thrift service.thrift
```

**Keep only two methods and their dependencies:**

```sh
trimmer -m UserService.GetUser -m UserService.DeleteUser -o slim.thrift service.thrift
```

**Trim an entire IDL tree recursively:**

```sh
trimmer -r ./idl -o ./trimmed_idl ./idl/service.thrift
```

The output directory must be outside the `-r` base directory to avoid overwriting source files.

**Force-trim everything, ignoring `@preserve` comments and config-based preservation:**

```sh
trimmer -p false -o service_slim.thrift service.thrift
```

**Use a config file to pin preserved structs:**

```sh
# trim_config.yaml in working directory is picked up automatically
trimmer -o slim.thrift service.thrift
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. |
| `2` | Error: invalid arguments, IDL parse failure, semantic error, or I/O error. |

## Troubleshooting

| Problem | Cause | Fix |
|---|---|---|
| `require exactly 1 argument for the IDL parameter` | No input file given, or more than one positional argument. | Pass exactly one `.thrift` file as the last argument. |
| `found include circle` | Circular `include` chain detected. | Remove the circular dependency from the IDL files. |
| `warning: method X not found in Y` | A `-m` pattern did not match any method in the IDL. | Check the spelling of `ServiceName.MethodName`. On a multi-service IDL, always include the service name prefix. |
| `-o should be set as a valid dir to enable -r` | `-r` was used and `-o` points to an existing file (not a directory). | Use a directory path for `-o`, or omit `-o` to use the default `./trimmed_idl`. |
| `output-dir should be set outside of -r base-dir to avoid overlay` | The `-o` directory is inside or equal to the `-r` base directory. | Move the output directory outside the IDL source tree. |
| A struct that should be removed is kept | The struct has a `// @preserve` comment, or is listed in `preserved_structs`/`preserved_files` in the config. | Pass `-p false` to disable all preservation and force removal. |
| An included file appears in the output even though none of its types are used | The included file has constants, enums, or typedefs — these are always kept. | This is expected behavior; constants/enums/typedefs are never removed. |
