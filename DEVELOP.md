# Develop docs

## Contributing policy

- This project accept AI code, but it's forborden to generate all code by AI.

## Layout

```text
mc-saver
├── cmd/               CLI, module acovia.net/mc-saver
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
├── save/              backup engine: config loading, world walking, zip writing
│   ├── save.go        current layout: dimensions/<namespace>/<id>/, plus saveDirInRoot
│   └── old_save.go    the `-l` layout: DIM-1 / DIM1 dimension roots
├── record/            logging helpers
├── build.sh           cross-compile the release targets
├── go.work            workspace tying the three modules together
├── README.md          user documentation
└── DEVELOP.md         this file
```

Build and run from the repo root:

```bash
go build ./cmd
./mc-saver help
```

## Command layer

```
mc-saver [-l] [command] <world> [args...]
```

## program

Program has two mode:

- cli mode: `main` prints the error with `record.Error` and stops.
- repl mode: `replLoop` prints it with `record.ErrorNoExit` and keeps the session.

This program use `subCmdArgs` slice to store arguments from command line, compared to the usual command usage, the REPL mode omits the world directory argument.

The program behaves like:

### Program runtime

Program will judge weather first argument is path, by judge weather start with `./` or `/`.

Program will search command function from map `cmdMap`, if command found in this map, program will execute corresponding function, if not found in map, program will throw a `unknown command` error.

Some commands will read config from `<world>/saver.json` file, program will set config = nil if this file not found. 

Program store subcommand arguments in slice `subCmdMap`, subcommand function will read arguments from this slice.

### Repl

Enter a repl, if you run with no arguments, program will ask world you want to selected, you neeed to run `select <world>` to select world, To run `mc-saver repl <world>`, or `mc-saver <world>` in CLI will enter repl with selected world, when selected, world path will be stored in `subCmdArgs[0]`, and will not reset unless run `select <world>` in repl.

### `list` commands

then call function to print config about you want to know.

### `add` commands

These commands will read config and add rule to config, then save config to `<world>/saver.json`.

### `del` commands

These commands will read config and del rule with index number, then save config to `<world>/saver.json`.

### `mod` commands

These commands will read config and modify rule with index number, then save config to `<world>/saver.json`.

### `run` commands

`run` command start a backup, if output is not specify, program will use default output path `./` to save archive, program will create a temp file, and return a writer contain zip writer, file object and output path string, writer point to a temp file in directory at the same level as output path.

Program will init output path with function `initOutputPath` before init writer, if target path is a directory, function will return a path named `<world_name-$time>` under directory.

Program will call `Package()` method of writer to package archive, `Package()` will rename temp file to output path, program will add subfix at after file name and before extention name.

### `gencfg` command

this command will generate a default config if config is empty or config not found, program will throw a error if config is not empty.