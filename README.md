# mc-saver

A Minecraft world backup tool written in Go. It reads a JSON rule file and packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

## Features

- **Selective backups** — back up only the dimensions and chunks you care about, so archives are smaller and faster than a full copy.
- **Per-dimension rules** — the overworld, the Nether, the End, and any custom dimension.
- **Two ways to pick regions** — `range` rectangles, or `simple` for individual regions.
- **Config managed from the command line** — add, modify, remove and list dimensions, rules and files without hand-editing JSON.
- **REPL mode** — select the world once, then run commands interactively; tab completion covers command names and dimension ids.

## Install

Download a prebuilt binary from the [Releases](https://github.com/fovlin/mc-saver/releases) page, or build from source — see [DEVELOP.md](DEVELOP.md). Release history lives in [CHANGELOG.md](CHANGELOG.md).

### Upgrading from 1.x

2.0.0 renames the config command families to match their Unix counterparts. The old names were removed, so update scripts before upgrading:

| 1.x | 2.0.0 |
| --- | --- |
| `list`, `list-config`, `list-dms`, `list-range`, `list-simple`, `list-file` | `ls`, `ls-config`, `ls-dms`, `ls-range`, `ls-simple`, `ls-file` |
| `del-dms`, `del-range`, `del-simple`, `del-file` | `rm-dms`, `rm-range`, `rm-simple`, `rm-file` |
| `mod-dms`, `mod-range`, `mod-simple`, `mod-file` | `set-dms`, `set-range`, `set-simple`, `set-file` |

Only the command names changed; the arguments and the `saver.json` format are unchanged. `ls*` only reads, `rm-*` removes an entry, and `set-*` replaces an entry (`set-dms` renames a dimension).

## Usage

```
mc-saver [-l] [command] <world> [args...]
```

Every command that operates on a world takes the world directory as its first argument; the rule file is read from `<world>/saver.json`. `help` and `about` take no world, and `repl` takes an optional one. Flags, if any, go before the command.

When the first argument looks like a path — it starts with `./` or `/`, after `\` is normalized to `/` — it is taken as the world and REPL mode starts straight away, so `mc-saver /srv/minecraft/world` means the same as `mc-saver repl /srv/minecraft/world`.

```bash
mc-saver gencfg /srv/minecraft/world              # write the default rule file
mc-saver run /srv/minecraft/world backup.zip      # back it up
mc-saver ls-dms /srv/minecraft/world              # inspect the rules
mc-saver add-simple /srv/minecraft/world minecraft:overworld @-3 @-3  # negative coords need @
mc-saver -l run /srv/minecraft/old-world out.zip  # legacy worlds (before 1.21.11)
```

A missing rule file is treated as an empty one, so every command works out of the box; the file appears on disk only once a config command or `gencfg` writes it. `gencfg <world>` writes the full default rule file — three dimensions (`overworld`, `the_nether`, `the_end`), each with the `-1,-1` to `0,0` range, plus four root entries (`level.dat`, `data`, `datapacks`, `players`).

### The `repl` mode

In REPL mode the world is chosen once and every following command uses it, so the world path is typed once instead of on every command line.

```bash
mc-saver repl </path/to/world>
```

```bash
# Enter REPL mode
root@mc-saver ~$ ./mc-saver repl /tmp/testworld
[INFO]: selected: '/tmp/testworld', type 'help' for commands, 'exit' to quit repl mode
mc-saver >
```

REPL mode puts the terminal in raw mode for line editing, so it needs a real terminal; piped input is rejected.

Type any [command](#commands) at the `mc-saver >` prompt. `mc-saver` without arguments enters REPL mode too, and asks for a world first (`select <world>`). `select <world>` switches worlds and prints `selected: <world>`; `world` prints the current selection without changing it. `Ctrl-D` or `exit` leaves REPL mode. Tab completes command names, and dimension ids while typing a dimension argument.

Run `help` inside the REPL for the full command list:

```bash
mc-saver > help
repl commands, the selected world is implicit:

	run [output]                back up the selected world
	gencfg                      write the default rule file
	ls                          list all dimension rules and file rules
	ls-config <dimension>...    list range and simple rules
	ls-dms                      list dimension namespace ids
	add-dms <dimension>...      add dimensions with the default range rule
	rm-dms <dimension>...       delete dimensions
	set-dms <old> <new>         rename a dimension, keeping its rules
	ls-range <dimension>...     list range rules
	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
	rm-range <dimension> <index>...
	set-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>
	ls-simple <dimension>...    list simple rules
	add-simple <dimension> <x> <y>
	rm-simple <dimension> <index>...
	set-simple <dimension> <index> <x> <y>
	ls-file                     list file rules
	add-file <name>...
	rm-file <index>...
	set-file <index> <name>

	world                       print the selected world
	select <world>              switch to another world
	help                        print this help text
	about                       print the information and copyright
	exit                        leave repl mode
mc-saver >

```

### Commands

Backup and utility:

| Command | Description |
| --- | --- |
| `run <world> [output]` | Back up a world. Output defaults to `.`; an existing directory receives `<world>-YYYY-MM-DD.zip`, and any other path is used as the file name. |
| `gencfg <world>` | Write the default rule file to `<world>/saver.json`. An empty rule file is filled with the defaults; one that already has rules is left alone and reported as an error. |
| `repl [world]` | Enter interactive REPL mode, for `[world]` if it is given. Running `mc-saver` without arguments does the same. |
| `help` | Print the built-in usage text. |
| `about` | Print the name, repository and license. |

The REPL adds four commands of its own:

| Command | Description |
| --- | --- |
| `world` | Print the selected world. |
| `select <world>` | Switch to another world; the current one is kept if the new one cannot be loaded. |
| `help` | Print the shorter REPL usage text. |
| `exit` | Leave repl mode. |

The config command families:

| Command | Description |
| --- | --- |
| `ls <world>` | Print every dimension rule, then the file rules. |
| `ls-config <world> <dimension>...` | Print both the range and the simple rules of a dimension. |
| `ls-dms <world>` | Print the dimension ids. |
| `add-dms <world> <dimension>...` | Add dimensions with the default range rule (`-1,-1` to `0,0`). |
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

A `<dimension>` is a namespace id like `minecraft:overworld`. `add-range` and `add-simple` create the dimension if it does not exist yet; the other dimension commands report an error instead.

The `add-*`, `rm-*` and `set-*` families write the rule file back after they run; `ls-*` only reads it.

On the command line, a negative coordinate must be written with a literal prefix, for example `@-3`, because `-3` would be parsed as an option. `@`, `:` and `=` are accepted; the prefix is stripped before parsing. In the REPL, plain `-3` works because the line parser does not treat it as an option.

Indices are 0-based, as printed by the `ls*` commands, and refer to the list as it is before the command runs. An index outside the list is an error; nothing is written in that case.

### Flags

| Flag | Description |
| --- | --- |
| `-l`, `--legacy` | Back up using the legacy single-folder layout (`DIM-1`/`DIM1`), for worlds from before 1.21.11. Also makes `gencfg` write a legacy-friendly `file` list. |
| `-c`, `--color` | Enable ANSI color output. |
| `-t`, `--time` | Prepend a timestamp to every log line. |
| `--help` | Print the help text and exit. |

Flags are switches and take no value. Put them before the command: `mc-saver -l run <world>`.

## Configuration

The rule file lives inside the world directory as `saver.json`, and has two top-level fields: `dimension` and `file`. Maintain it with the commands above — hand-editing is **not recommended**, because a misspelled field name is silently ignored.

```json
{
	"dimension": {
		"minecraft:overworld": {
			"range": [
				{
					"from": {
						"x": -1,
						"y": -1
					},
					"to": {
						"x": 0,
						"y": 0
					}
				}
			],
			"simple": [
				{
					"x": 8,
					"y": 8
				}
			]
		},
		"minecraft:the_end": {
			"range": [
				{
					"from": {
						"x": -1,
						"y": -1
					},
					"to": {
						"x": 0,
						"y": 0
					}
				}
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

### `dimension`

Keyed by dimension namespace id. Each dimension takes either or both of:

- **`range`** — rectangles, inclusive: every region between `from` and `to` is backed up. The two corners may be given in any order.
- **`simple`** — individual regions, for the few that fall outside your rectangles.

`x`/`y` are Minecraft's region indices (`r.<x>.<y>.mca`); one region covers 512×512 blocks.

### `file`

Files or folders at the world root. Files are added as-is; folders are walked recursively.

## Notes

- A missing `<world>/saver.json` counts as an **empty** rule file — no dimension and no file rules. That keeps the `ls*` commands usable, and `gencfg <world>` later writes the defaults (it refuses to touch a file that already has rules). The file itself is written only when a config command or `gencfg` runs, so a world can be inspected without leaving anything behind. A backup with no rules produces an archive with nothing in it and still reports success, so watch the archive size.
- Everything named in the rule file must exist in the world: a configured dimension without its `dimensions/<namespace>/<id>` directory, or a `file` entry that is missing, aborts the whole backup instead of being skipped. That is deliberate — otherwise a typo or a stale rule would silently produce an incomplete archive. Remove what you don't have with `rm-dms` / `rm-file`.
- Missing *region* files are a different matter: a region that was never generated — or a whole missing `region` directory inside an existing dimension — is skipped with a warning, because that is world content rather than something the rule file declares.
- A region listed more than once — by overlapping `range` rules, or by a `range` and a `simple` — is added to the archive once per occurrence.
- Legacy worlds: `gencfg -l <world>` writes a legacy-friendly `file` list (`playerdata` instead of `players`, plus `advancements`). That list also carries `data`, which in `-l` mode is the overworld's own data folder, so its files are stored twice; drop the `data` entry with `rm-file` if that matters.
- The archive keeps the world directory name as its top-level folder: backing up `/srv/minecraft/world` stores `world/level.dat`, `world/dimensions/...` and so on, so extracting it under the server directory puts the world back in place.
- Output naming: an existing directory receives `<world>-YYYY-MM-DD.zip`; if that name is taken, `-1`, `-2`, … is appended before the last extension. A path that does not exist is used as the archive file name as-is, and its parent directories are created.
- Progress lines (`INFO`) go to stdout, while `WARN` and `ERROR` go to stderr, so `mc-saver ... 2>/dev/null` keeps the progress and hides the diagnostics.
- Symbolic links are **not supported**. The walk does not follow them, and a link that points at a directory or outside the world aborts the whole backup, as does any other entry the rule file cannot read. A link to a plain file inside the world is stored under the link's name with the target's content. Replace links with real directories, or list the entries you need as `file` rules instead of the folder that contains the link.
