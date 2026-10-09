# Develop docs

## Contributing policy

The projects of fovlin are all handwritten, and all projects reject AI coding.
It is strictly forbidden to use AI audit results to submit as issues.
If you can't accept it, please leave.

## Layout

```text
mc-saver
├── cmd/          CLI, module acovia.net/mc-saver
│   ├── main.go        entry point, flag parsing, the shared command table
│   ├── repl.go        interactive mode and its own command names
│   ├── data.go        shared state, message builders, pure helpers
│   ├── init.go        world and rule file loading
│   ├── help.go        both usage texts
│   ├── io.go          zip writing and rule file writing
│   ├── run.go         run command
│   ├── gencfg.go      gencfg command
│   ├── list.go        list and list-config commands
│   ├── dimension.go   the dms command family
│   ├── range.go       the range command family
│   ├── simple.go      the simple command family
│   ├── file.go        the file command family
│   └── about.go       about command
├── save/         backup engine: config loading, world walking, zip writing
│   ├── save.go        current layout: dimensions/<namespace>/<id>/, plus saveDirInRoot
│   └── old_save.go    the `-l` layout: DIM-1 / DIM1 dimension roots
├── record/       logging helpers
├── build.sh      cross-compile the release targets
├── go.work       workspace tying the three modules together
├── README.md     user documentation
└── DEVELOP.md    this file
```

Build and run from the repo root:

```bash
go build -o mc-saver ./cmd
./mc-saver help
```

Patterns resolve per module: use `go build ./cmd` or `go vet ./cmd/...`. A bare
`./...` from the root does not match anything, because the root is not itself a module.

## Command layer

```
mc-saver [-l] [-color] <command> <world> [args...]
```

Every command has the signature `func() error`. Two package-level values carry its
input: `subCmdArgs` (`[world, args...]`) and `config`. A command never reports or exits
by itself, the caller decides:

- cli mode: `main` prints the error with `record.Error` and stops.
- repl mode: `replLoop` prints it with `record.ErrorNoExit` and keeps the session.

`initProgram` fills `cmd` and `subCmdArgs`, and rewrites the arguments into `repl <path>`
when the first one looks like a path (`./...` or `/...`). The repl builds the same
`subCmdArgs` from the input line, which is what lets both modes share one implementation
of every command.

The world directory and the rule file path are **not** package-level: a command derives
them with `fmtPath` and passes them down.

```go
worldDirPath := fmtPath(subCmdArgs[0])
configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
	return err
}
```

`initWorldConfig(worldDirPath, configFilePath)` validates the world and loads the rule
file; `initConfig` creates an empty one when it is missing. Commands that change rules
write them back with `commitConfig(configFilePath)`.

`cmdMap` holds the commands shared by both modes. `replOnlyCmd` adds `world`, `select` and
`exit`, and replaces `help` with the shorter repl text. Both tables are filled in `init`
because `repl` reads them back, which would otherwise be an initialization cycle.

Malformed invocations go through `usageError(argsLine, replArgsLine)`: both usage lines
are written out per command, one for cli mode (`mc-saver run <world> [output]`) and one
for repl mode (`run [output]`), so the two modes still share a single implementation.

## Save engine

`run` hands the engine an `*os.Root` of the world, the loaded config, a `*zip.Writer`, and
a `saveFile` callback. Every path the engine passes back is relative to that root, so the
callback is what maps it onto a zip entry — this is also what keeps a rule file from
reading outside the world. Two entry points: `SaveAllFile` for the current layout and
`SaveOldAllFile` for `-l`.

Both walk the same three things: the configured regions, the `data` subdirectory of each
dimension, and every `file` rule. Region paths are built by `FormatRegionFilePath` from a
`RangeIterator` (`range` rules) or one `Coordinate` (`simple` rules). Directory recursion
is shared: `saveDirInRoot` opens the subdirectory as a root of its own, walks it, and
closes it before returning — closing per call is deliberate, `defer` inside the caller's
loop would hold one handle per configured directory until the loop ended. Its result
wraps the walk error, not the `Close` error.

`SaveRootDataFile` is shared by both entry points; the two differ only in how
`SaveDimensionFile` / `SaveOldDimensionFile` resolve the dimension root. The engine
reports nothing itself, it only returns errors — the zip writer's `Close` and the removal
of a partially written archive on failure both belong to `run`.

## Release

1. `bash build.sh` — compiles 6 targets (linux / darwin / windows × amd64 / arm64) and
   leaves one `.tar.gz` per target in `build/`.
2. `git tag vX.Y.Z && git push origin vX.Y.Z`
3. Attach the six `build/*.tar.gz` files to the GitHub release.

There is no automated test suite yet; smoke-test the commands by hand before tagging.

## Conventions

- Commands return an error instead of printing it; only the cli entry point and the repl
  loop report. Nothing below `main` exits: `record.Error` is called from `main` alone and
  is the only thing that exits with a failure status, `exit` is the only other `os.Exit`.
- User-facing messages (`record.Error` / `Info` / `Warn`, `fmt.Print*`) carry no trailing
  period. The context leads and the cause follows after a colon: `record.Error("load config:", err)`.
- Returned errors are lowercase, carry no punctuation, and wrap the cause with `%w`:
  `fmt.Errorf("open file: %w", err)`. When a layer has context of its own it leads the same
  way (`fmt.Errorf("save directory: %w", err)`), so a reader sees the outermost context
  first and the root cause last.
- Version tags use a `v` prefix (`v1.2.0`).
- Commands validate their arguments before loading the config, and validate every index
  before touching it: an out-of-range index aborts the command and leaves the file unchanged.

## Not a bug

Intentional behaviours, in case they look like defects:

- A missing rule file is created **empty** (no dimension and no file rules), so a world can
  be inspected without a preparation step. `gencfg <world>` fills that empty file with the
  defaults and refuses to touch one that already has rules ("config not empty").
- Everything named in the rule file must exist in the world. A configured dimension whose
  `dimensions/<namespace>/<id>` directory is missing, or a `file` entry that is missing,
  aborts the whole backup instead of being skipped — silently skipping would hide a stale
  config or a typo. Missing *region* files are skipped with a warning, and so is a whole
  missing `region` directory inside an existing dimension, because those are world content
  rather than declared input.
- Missing world content is skipped silently rather than reported: a dimension without a
  `data` subdirectory, or an entry that disappears between the walk and the read, never
  fails the backup. Only input declared in the rule file is treated as a contract.
- A rule file with no rules backs up nothing: the archive is created, holds zero entries and
  the run still reports success.
- `add-range` and `add-simple` create a dimension that is not configured yet; the other
  dimension commands report an error instead. `del-dms` only warns when a dimension is absent.
- A `range` whose `from` corner is larger than its `to` corner is not an error: the
  rectangle is read as the bounding box of the two corners, so `5,5` to `1,1` covers the
  same regions as `1,1` to `5,5`.
- `-l` and `-color` are switches; `-l=false` is not part of the interface.
- `-l` selects the dimension roots the backup reads: `.` for the overworld, `DIM-1` for the
  nether, `DIM1` for the end, and `dimensions/<namespace>/<id>` for anything else. Region
  files and the `data` subdirectory both hang off that root, so legacy mode backs up
  `DIM-1/data`, while the default mode backs up `dimensions/minecraft/the_nether/data`. The
  rule file format is the same in both modes.
- The config file format is not backward compatible — regenerate it with `gencfg` after
  upgrading.
- A region listed more than once in the config (duplicate or overlapping rules, or a
  `range` and a `simple` covering the same region) is added to the archive once per
  occurrence, so the archive can contain duplicate entries.
- If the output name is already taken, the first free `<name>-<n><ext>` is used instead of
  failing or overwriting: starting at `-0`, before the extension, so `bk/backup.zip` becomes
  `bk/backup-0.zip`. The directory part is never touched.
