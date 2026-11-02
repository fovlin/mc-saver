# mc-saver

Back up selected parts of a Minecraft world into a ZIP archive.

## Install

Prebuilt binaries are on the [releases page](https://github.com/fovlin/mc-saver/releases). To build from source, see [DEVELOP.md](DEVELOP.md).

## Usage

```
mc-saver [-l] [command] <world> [args...]
```

The world directory comes first for every command except `help` and `about`. Rules are read from `<world>/saver.json`. A first argument that looks like a path (`./` or `/`) enters REPL mode with that world.

```bash
mc-saver gencfg /srv/minecraft/world              # write the default rule file
mc-saver run /srv/minecraft/world backup.zip      # back it up
mc-saver ls /srv/minecraft/world                  # inspect the rules
mc-saver -l run /srv/minecraft/old-world out.zip  # world from before 1.21.11
```

On the command line, prefix a negative coordinate with `@`, `:` or `=`, as in `@-3`; otherwise `-3` is read as an option. Inside the REPL the plain number works.

## REPL

```bash
mc-saver repl /srv/minecraft/world
mc-saver /srv/minecraft/world
mc-saver ./world
```

The world is selected once, then every command is typed at the `mc-saver >` prompt without the world argument. `select <world>` switches worlds, `world` prints the current one, `help` lists the commands, `Ctrl-D` or `exit` leaves. Tab completes command names and dimension ids. It needs a real terminal.

## Commands

| Command | Description |
| --- | --- |
| `run <world> [output]` | Back up a world. Defaults to `.`. |
| `gencfg <world>` | Write the default rule file. |
| `repl [world]` | Enter REPL mode. |
| `help`, `about` | Print the usage text, or the version and license. |
| `ls <world>` | List every dimension rule and file rule. |
| `ls-config <world> <dimension>...` | List the range and simple rules of a dimension. |
| `ls-dms <world>` | List the dimension ids. |
| `add-dms <world> <dimension>...` | Add dimensions with the default range rule. |
| `rm-dms <world> <dimension>...` | Delete dimensions. |
| `set-dms <world> <old> <new>` | Rename a dimension, keeping its rules. |
| `ls-range <world> <dimension>...` | List the range rules, with indices. |
| `add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>` | Append a range rule. |
| `rm-range <world> <dimension> <index>...` | Delete range rules. |
| `set-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>` | Replace a range rule. |
| `ls-simple <world> <dimension>...` | List the simple rules, with indices. |
| `add-simple <world> <dimension> <x> <y>` | Append a single region. |
| `rm-simple <world> <dimension> <index>...` | Delete simple rules. |
| `set-simple <world> <dimension> <index> <x> <y>` | Replace a simple rule. |
| `ls-file <world>` | List the file rules, with indices. |
| `add-file <world> <name>...` | Append file rules. |
| `rm-file <world> <index>...` | Delete file rules. |
| `set-file <world> <index> <name>` | Replace a file rule. |

A `<dimension>` is a namespace id like `minecraft:overworld`. Indices are 0-based.

## Flags

| Flag | Description |
| --- | --- |
| `-l`, `--legacy` | Read the legacy layout (`DIM-1`/`DIM1`). |
| `-c`, `--color` | Colored output. |
| `-t`, `--time` | Timestamp on every log line. |
| `--help` | Print the help text. |

## saver.json

```json
{
	"dimension": {
		"minecraft:overworld": {
			"range": [
				{ "from": { "x": -1, "y": -1 }, "to": { "x": 0, "y": 0 } }
			],
			"simple": [
				{ "x": 8, "y": 8 }
			]
		}
	},
	"file": [
		"level.dat",
		"data",
		"datapacks",
		"players"
	]
}
```

`dimension` is keyed by namespace id; `range` holds inclusive rectangles whose corners may be given in any order, `simple` holds single regions. `x`/`y` are region indices, one region being 512×512 blocks. `file` lists files or folders at the world root. Use the commands above rather than editing the file.

## Upgrading

2.0.0 renamed the config command families: `list*` to `ls*`, `del-*` to `rm-*`, `mod-*` to `set-*`. Arguments and the file format are unchanged. See [CHANGELOG.md](CHANGELOG.md).

How a missing rule file, a missing region or a symlink is handled is described in [DEVELOP.md](DEVELOP.md).
