# Develop docs

## Contributing policy

- AI-generated code is accepted, but generating the whole project with AI is not allowed.

## Layout

```text
mc-saver
├── cmd/               CLI, module acovia.net/mc-saver
│   ├── main.go        entry point, flag parsing, cmdInfo/cmdMap
│   ├── repl.go        interactive mode, dispatch, tab completion
│   ├── init.go        shared state, world/config loading, helpers
│   ├── help.go        both usage texts
│   ├── io.go          zip writing and rule file writing
│   ├── run.go         run command
│   ├── gencfg.go      gencfg command
│   ├── list.go        ls and ls-config commands
│   ├── dms.go         the dms command family
│   ├── range.go       the range command family
│   ├── simple.go      the simple command family
│   ├── file.go        the file command family
│   └── about.go       about command
├── save/              backup engine: config loading, world walking, zip writing
│   ├── save.go        current layout: dimensions/<namespace>/<id>/, plus namespace listing
│   └── old_save.go    the `-l` layout: DIM-1 / DIM1 dimension roots
├── opt/               POSIX-style option parser, module acovia.net/opt
├── record/            logging helpers, plus the terminal (raw mode) flag
├── build.sh           cross-compile the release targets
├── go.work            workspace tying the four modules together
├── CHANGELOG.md       release history
├── README.md          user documentation
└── DEVELOP.md         this file
```

Build and run from the repo root:

```bash
go build ./cmd
./mc-saver help
```

## Command layer

```text
mc-saver [-l] [command] <world> [args...]
```

Commands are registered once in `cmdInfo`:

```go
type CommandObject struct {
	MinArgsNumber    int
	NeedDimensionArg bool
	Function         func() error
}
```

| Field | Meaning |
| --- | --- |
| `MinArgsNumber` | Minimum length of `subCmdArgs`, including the world at index 0. |
| `NeedDimensionArg` | Whether the command takes a dimension id at index 1; the REPL consults it before completing a dimension id. |
| `Function` | The command handler. |

`initProgram` derives `cmdMap` from `cmdInfo`, so `main` and the REPL resolve handlers from the same table. Add a new command to `cmdInfo`, not to `cmdMap`.

## Program flow

The program has two modes:

- CLI mode: `main` prints errors with `record.Error` and exits.
- REPL mode: `replLoop` prints errors with `record.ErrorNoExit` and keeps the session.

`subCmdArgs` holds the command arguments. In CLI mode `subCmdArgs[0]` is the world; in REPL mode it is the selected world, and the remaining elements are the words typed at the prompt.

When the first argument looks like a path — it starts with `./` or `/` after `\` is normalized to `/` — the program treats it as the world and enters REPL mode directly.

`main` looks up the command in `cmdMap`; an unknown name is an `unknown command` error. Commands that read a rule file use `save.NewNullConfig()` when `<world>/saver.json` does not exist.

Commands that read or write the rule file start with `loadWorld(minArgs, usage, replUsage)`: it checks the argument count, resolves the world path with `filepath.Abs` and `fmtPath`, and loads the rule file. `run` uses the same helper, so every command resolves the world the same way. `gencfg` and `repl` only validate the world with `initWorld`, because they must not load a rule file.

Command handlers read their arguments from `subCmdArgs`; they do not receive them as parameters.

### REPL

Running with no command enters REPL mode and asks for a world (`select <world>`). `mc-saver repl <world>` and `mc-saver <path>` enter REPL mode with the world already selected. The selected world stays in `subCmdArgs[0]` until `select <world>` changes it.

`replExec` splits each line with `opt.ParseLine`, then checks `replOnlyCmd` before `cmdMap`. Tab completion completes command names from `cmdMap`, and dimension ids from `<world>/dimensions/` for the second word.

### `ls` commands

`ls`, `ls-config`, `ls-dms`, `ls-range`, `ls-simple` and `ls-file` read the loaded config and print it. They do not write `saver.json`.

### `add` commands

These commands read the config, append a dimension, range, simple rule or file entry, and write the config back to `<world>/saver.json`.

### `rm` commands

These commands read the config, remove the rules at the given 0-based indices, and write the config back to `<world>/saver.json`.

### `set` commands

These commands read the config, replace the rule at the given 0-based index (or rename a dimension for `set-dms`), and write the config back to `<world>/saver.json`.

### Coordinates

Range and simple arguments are integers parsed by `convertIntArray`, which strips a leading `@`, `:` or `=` from each value. On the command line, write `@-3` for a negative coordinate because `opt` would otherwise read `-3` as an option. In the REPL the prefix is optional, because `opt.ParseLine` only splits words.

### `run` command

`run` starts a backup. If no output is given, it uses the default output path `.`. It creates a temporary file in the output directory and returns a writer containing the zip writer, the file object and the output path.

Before creating the writer, `run` resolves the output path with `initOutputPath`. If the target is a directory, the function returns `<world_name>-YYYY-MM-DD.zip` inside that directory.

`Package()` renames the temporary file to the output path. If that name is already taken, it adds `-1`, `-2`, … after the file name and before the extension. If any step fails, `run` calls `Clear()`, which closes the writer and removes the temporary file. `Writer.Close` is idempotent, so `Clear` still works when `Package` has already closed the writer before failing.

### `gencfg` command

`gencfg` writes the default rule file when the config is empty or missing, and reports an error when the config already has rules.

## Release

1. Update `CHANGELOG.md` and the migration notes in `README.md`.
2. Run `gofmt -l`, `go vet ./cmd`, and build each module (`cmd`, `opt`, `record`, `save`).
3. Smoke-test `run`, `gencfg`, `ls*`, `rm*` and `set*` against a scratch world.
4. Commit, merge into `main`, tag `vX.Y.Z`, and push the commit and tag.
5. Run `./build.sh` and attach the archives under `build/` to the release.

The project follows [Semantic Versioning](https://semver.org/). Breaking command-line changes bump the major version; the 2.0.0 command rename is one.
