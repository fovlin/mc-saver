# mc-saver

A Minecraft world backup tool written in Go. It reads a JSON rule file and packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

## Features

- **Selective backups** — back up only the dimensions and chunks you care about.
- **Per-dimension rules** — the overworld, the Nether, the End, and any custom dimension.
- **Two ways to pick regions** — `range` rectangles, or `simple` for individual regions.
- **Config managed from the command line** — no hand-editing JSON.
- **REPL mode** — select the world once, then work interactively; tab completes command names and dimension ids.

## Install

Download a prebuilt binary from the [Releases](https://github.com/fovlin/mc-saver/releases) page, or build from source — see [DEVELOP.md](DEVELOP.md). Release history lives in [CHANGELOG.md](CHANGELOG.md).

### Upgrading from 1.x

2.0.0 renames the config command families to match their Unix counterparts. The old names were removed, so update scripts before upgrading:

| 1.x | 2.0.0 |
| --- | --- |
| `list`, `list-config`, `list-dms`, `list-range`, `list-simple`, `list-file` | `ls`, `ls-config`, `ls-dms`, `ls-range`, `ls-simple`, `ls-file` |
| `del-dms`, `del-range`, `del-simple`, `del-file` | `rm-dms`, `rm-range`, `rm-simple`, `rm-file` |
| `mod-dms`, `mod-range`, `mod-simple`, `mod-file` | `set-dms`, `set-range`, `set-simple`, `set-file` |

Only the command names changed; the arguments and the `saver.json` format are unchanged.

## Usage

```
mc-saver [-l] [command] <world> [args...]
```

Every command that operates on a world takes the world directory as its first argument; the rule file is read from `<world>/saver.json`. `help` and `about` take no world, and `repl` takes an optional one. Flags go before the command.

```bash
mc-saver gencfg /srv/minecraft/world              # write the default rule file
mc-saver run /srv/minecraft/world backup.zip      # back it up
mc-saver ls-dms /srv/minecraft/world              # inspect the rules
mc-saver add-simple /srv/minecraft/world minecraft:overworld @-3 @-3  # negative coords need @
mc-saver -l run /srv/minecraft/old-world out.zip  # legacy worlds (before 1.21.11)
```

On the command line a negative coordinate must be written with a literal prefix, for example `@-3`, because `-3` would be parsed as an option. `@`, `:` and `=` are accepted, and the prefix is stripped. In the REPL, plain `-3` works.

### The `repl` mode

Select the world once, then run commands without repeating the path. Running `mc-saver` with no arguments enters REPL mode and asks for a world first.

```bash
mc-saver repl </path/to/world>
```

```text
root@mc-saver ~$ ./mc-saver repl /tmp/testworld
[INFO]: selected: '/tmp/testworld', type 'help' for commands, 'exit' to quit repl mode
mc-saver > run
```

A path as the first argument does the same: `mc-saver /srv/minecraft/world`. REPL mode puts the terminal in raw mode for line editing, so it needs a real terminal. `select <world>` switches worlds, `world` prints the current one, `help` prints the command list, and `Ctrl-D` or `exit` leaves. Tab completes command names, and dimension ids while typing a dimension argument.

### Commands

Backup and utility:

| Command | Description |
| --- | --- |
| `run <world> [output]` | Back up a world. Defaults to `.`. |
| `gencfg <world>` | Write the default rule file to `<world>/saver.json`. |
| `repl [world]` | Enter interactive REPL mode. |
| `help` | Print the built-in usage text. |
| `about` | Print the name, repository and license. |

The REPL adds four commands of its own: `world`, `select <world>`, `help`, and `exit`.

The config command families:

| Command | Description |
| --- | --- |
| `ls <world>` | Print every dimension rule, then the file rules. |
| `ls-config <world> <dimension>...` | Print both the range and the simple rules of a dimension. |
| `ls-dms <world>` | Print the dimension ids. |
| `add-dms <world> <dimension>...` | Add dimensions with the default range rule. |
| `rm-dms <world> <dimension>...` | Delete dimensions. |
| `set-dms <world> <old> <new>` | Rename a dimension, keeping its rules. |
| `ls-range <world> <dimension>...` | Print the range rules, with their indices. |
| `add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>` | Append a rectangular range rule. |
| `rm-range <world> <dimension> <index>...` | Delete the range rules at `<index>...`. |
| `set-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>` | Replace the range rule at `<index>`. |
| `ls-simple <world> <dimension>...` | Print the simple rules, with their indices. |
| `add-simple <world> <dimension> <x> <y>` | Append a single region. |
| `rm-simple <world> <dimension> <index>...` | Delete the simple rules at `<index>...`. |
| `set-simple <world> <dimension> <index> <x> <y>` | Replace the simple rule at `<index>`. |
| `ls-file <world>` | Print the file rules, with their indices. |
| `add-file <world> <name>...` | Append one or more file rules. |
| `rm-file <world> <index>...` | Delete the file rules at `<index>...`. |
| `set-file <world> <index> <name>` | Replace the file rule at `<index>`. |

A `<dimension>` is a namespace id like `minecraft:overworld`. Indices are 0-based, as printed by the `ls*` commands.

### Flags

| Flag | Description |
| --- | --- |
| `-l`, `--legacy` | Use the legacy single-folder layout (`DIM-1`/`DIM1`), for worlds from before 1.21.11. |
| `-c`, `--color` | Enable ANSI color output. |
| `-t`, `--time` | Prepend a timestamp to every log line. |
| `--help` | Print the help text and exit. |

Flags are switches and take no value: `mc-saver -l run <world>`.

## Configuration

The rule file lives inside the world directory as `saver.json`, with two top-level fields: `dimension` and `file`. Maintain it with the commands above rather than by hand.

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

`dimension` is keyed by dimension namespace id, and each entry takes either or both of:

- **`range`** — rectangles, inclusive; the two corners may be given in any order.
- **`simple`** — individual regions.

`x`/`y` are Minecraft's region indices (`r.<x>.<y>.mca`); one region covers 512×512 blocks.

`file` lists files or folders at the world root. Files are added as-is, folders are walked recursively.

## More

How a missing rule file, a missing region, a symlink or a taken output name is handled is documented in [DEVELOP.md](DEVELOP.md).
