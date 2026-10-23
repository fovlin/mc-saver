# opt

A POSIX-style command-line option parser for Go.

The standard `flag` package follows a `-flag value` convention that feels awkward if you spend your day in a Unix shell: short options have to be spelled out, `-abc` cannot be combined, and long options need separate registration. `opt` was built for exactly those irritations — it lets you write `-v`, `--verbose`, `-abc`, and `--key=value`, the same way you type any other command in a terminal.

## Features

- **Short and long options** — `-v` and `--verbose` can bind to the same variable, registered in one call
- **Combinable short options** — `-abc` is the same as `-a -b -c`
- **`--key=value` support** — equals-form long options are handled directly
- **Boolean switches never swallow arguments** — the token after `-v` is never mistaken for its value; it goes to `Args`
- **Boolean switches can trigger functions** — `BoolFunc` binds an option to an action
- **Zero dependencies** — standard library only

## Install

```bash
go get acovia.net/opt
```

> If you want to reference this repository directly, the module path in `go.mod` is `opt`; use a local `replace` directive.

## Quick start

```go
package main

import (
	"fmt"
	"os"

	"acovia.net/opt"
)

func main() {
	var (
		verbose bool
		output  string
		name    string
	)

	opt.BoolSwitch(&verbose, "v", "verbose")
	opt.StringVar(&output, "o", "output", "out.bin")
	opt.StringVar(&name, "", "name", "") // long option only

	opt.BoolFunc("h", "help", func() {
		fmt.Println("usage: app [-v] [-o dir] [--name n] files...")
		os.Exit(0)
	})

	if err := opt.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("verbose:", verbose)
	fmt.Println("output :", output)
	fmt.Println("name   :", name)
	fmt.Println("args   :", opt.Args)
}
```

```bash
$ app -v -o dist --name demo a.txt b.txt
verbose: true
output : dist
name   : demo
args   : [a.txt b.txt]
```

## API

### Registration

| Function | Description |
| --- | --- |
| `StringVar(&v, short, long, def)` | Register a string option with both short and long names |
| `ShortStringVar(&v, name, def)` | Register a short option only (`-name`) |
| `LongStringVar(&v, name, def)` | Register a long option only (`--name`) |
| `BoolSwitch(&v, short, long)` | Register a boolean switch with both short and long names |
| `ShortBoolSwitch(&v, name)` | Register a short boolean switch only (`-name`) |
| `LongBoolSwitch(&v, name)` | Register a long boolean switch only (`--name`) |
| `BoolFunc(short, long, fn)` | Register a boolean switch that calls `fn` when set |

### Parsing

| Function | Description |
| --- | --- |
| `Parse() error` | Parse `os.Args`, collect positional arguments into `Args`, and run the callbacks registered by `BoolFunc` |

### Variables

| Variable | Description |
| --- | --- |
| `Args` | Positional arguments remaining after `Parse()` |
| `Options` | Table of registered options (`map[string]Value`) |

## Differences from the standard `flag` package

Switch if these help; otherwise there is no reason to migrate:

| Behavior | `flag` | `opt` |
| --- | --- | --- |
| Combinable short options | No | Yes, `-abc` |
| Short / long names | Registered separately | Both in one call |
| `--key=value` | Yes | Yes |
| `--flag=false` | Yes | **No** — a boolean switch is `true` whenever present |
| Single `-` | Treated as an argument | Treated as an argument |
| Function-style switch | Build it yourself | `BoolFunc` |

## Design notes

A few deliberate trade-offs, documented here so they are not mistaken for oversights:

- **`Parse()` parses once.** Like the standard `flag` package, the option table and callback list are package-level state that `Parse()` clears at the end. Calling it repeatedly in one process is outside the design.
- **Boolean switches take no value.** `BoolFlag.SetValue` consumes no argument and exists only to satisfy the `Value` interface — a boolean switch has no "value", so arguments never pass through it.
- **The program name is not parsed.** Scanning starts at `os.Args[1:]`, so the executable path never leaks into `Args`.
- **Unknown options are an error.** No silent ignoring — a typo fails loudly instead of looking like it worked.

## License

See the LICENSE file in this repository.
