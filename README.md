# mc-saver

A Minecraft world backup tool written in Go. It reads a JSON rule file and packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

## Features

- **Selective backups** — back up only the dimensions and chunks you care about, so archives are smaller and faster than a full copy.
- **Per-dimension rules** — the overworld, the Nether, the End, and any custom dimension.
- **Two ways to pick regions** — `range` rectangles, or `simple` for individual regions.
- **Config managed from the command line** — add, modify, delete and list dimensions, rules and files without hand-editing JSON.
- **Repl Mode for easy using** - `repl` Start once, enter the command continuously, and enter the world path only once.

## Install

Download a prebuilt binary from the [Releases](https://github.com/fovlin/mc-saver/releases) page, or build from source — see [DEVELOP.md](DEVELOP.md).

## Usage

```
mc-saver [-l] [-color] [command] <world> [args...]
```

Every command takes the world directory as its first argument; the rule file is read from `<world>/saver.json`. Flags, if any, go before the command.

When the first argument looks like a path — it starts with `./` or `/` — it is taken as the world and repl mode starts straight away, so `mc-saver /srv/minecraft/world` means the same as `mc-saver repl /srv/minecraft/world`.

```bash
mc-saver gencfg /srv/minecraft/world           # write the default rule file
mc-saver run /srv/minecraft/world backup.zip   # back it up
mc-saver list-dms /srv/minecraft/world         # inspect the rules
mc-saver -l run /srv/minecraft/old-world out.zip   # legacy worlds (before 1.21.11)
```

A missing rule file is created empty on first use, so every command works out of the box. `gencfg <world>` writes the full default rule file instead — three dimensions (`overworld`, `the_nether`, `the_end`), each with the `-1,-1` to `0,0` range, plus four root entries (`level.dat`, `data`, `datapacks`, `players`).

### New Feature （The `repl` mode）
Now we have updated the repl mode for easy using. In this mode, users only need to enter the world path once to continuously input commands, which greatly reduces the annoyance of command errors caused by frequent CLI command input. below is how to use `repl` to execute your commands : 

- The format for `repl`mode : 
```zsh 
mc-saver repl </path/to/world>
```
- Example :

```bash 
# Enter Repl mode 
root@mc-saver ~$ ./mc-saver repl /tmp/testworld 

[2026-10-08 12:51:28 INFO]: selected: /tmp/testworld
[2026-10-08 12:51:28 INFO]: type 'help' for commands, 'exit' to quit repl mode
mc-saver >
```
In above example, you can type any [commands](#commands) when the `mc-saver >` appear. `mc-saver` without any argument enters repl mode too, and then asks for a world (`select <world>`).

If you wanna more information about how to use `repl` please check the `help` menu by below command :

```bash 
mc-saver > help
repl commands, the selected world is implicit:

	run [output]                back up the selected world
	gencfg                      write the default rule file
	list                        list all dimension rules and file rules
	list-config <dimension>...  list range and simple rules
	list-dms                    list dimension namespace ids
	add-dms <dimension>...      add dimensions with the default range rule
	del-dms <dimension>...      delete dimensions
	mod-dms <old> <new>         rename a dimension, keeping its rules
	list-range <dimension>...   list range rules
	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
	del-range <dimension> <index>...
	mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>
	list-simple <dimension>...  list simple rules
	add-simple <dimension> <x> <y>
	del-simple <dimension> <index>...
	mod-simple <dimension> <index> <x> <y>
	list-file                   list file rules
	add-file <name>...
	del-file <index>...
	mod-file <index> <name>

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
| `run <world> [output]` | Back up a world. Output defaults to `.`, where a dated zip is created. |
| `gencfg <world>` | Write the default rule file to `<world>/saver.json`. An empty rule file is filled with the defaults; one that already has rules is left alone and reported as an error. |
| `repl [world]` | Enter interactive repl mode, for `[world]` if it is given. Running `mc-saver` without arguments does the same. |
| `help` | Print the built-in usage text. |
| `about` | Print the name, repository and license. |

The repl adds three commands of its own:

| Command | Description |
| --- | --- |
| `select <world>` | Switch to another world; the current one is kept if the new one cannot be loaded. |
| `help` | Print the shorter repl usage text. |
| `exit` | Leave repl mode. |

Config commands write the rule file back after they run:

| Command | Description |
| --- | --- |
| `list <world>` | Print every dimension rule, then the file rules. |
| `list-config <world> <dimension>...` | Print both the range and the simple rules of a dimension. |
| `list-dms <world>` | Print the dimension ids. |
| `add-dms <world> <dimension>...` | Add dimensions with the default range rule (`-1,-1` to `0,0`). |
| `del-dms <world> <dimension>...` | Delete dimensions. |
| `mod-dms <world> <old> <new>` | Rename a dimension, keeping its rules. |
| `list-range <world> <dimension>...` | Print the range rules, with their indices. |
| `add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>` | Append a rectangular range rule. |
| `del-range <world> <dimension> <index>...` | Delete the range rules at `<index>...`. |
| `mod-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>` | Replace the range rule at `<index>`. |
| `list-simple <world> <dimension>...` | Print the simple rules, with their indices. |
| `add-simple <world> <dimension> <x> <y>` | Append a single region. |
| `del-simple <world> <dimension> <index>...` | Delete the simple rules at `<index>...`. |
| `mod-simple <world> <dimension> <index> <x> <y>` | Replace the simple rule at `<index>`. |
| `list-file <world>` | Print the file rules, with their indices. |
| `add-file <world> <name>...` | Append one or more file rules. |
| `del-file <world> <index>...` | Delete the file rules at `<index>...`. |
| `mod-file <world> <index> <name>` | Replace the file rule at `<index>`. |

A `<dimension>` is a namespace id like `minecraft:overworld`. `add-range` and `add-simple` create the dimension if it does not exist yet; the other dimension commands report an error instead.

Indices are 0-based, as printed by the `list*` commands, and refer to the list as it is before the command runs. An index outside the list is an error; nothing is written in that case.

### Flags

| Flag | Description |
| --- | --- |
| `-l` | Back up using the legacy single-folder layout (`DIM-1`/`DIM1`), for worlds from before 1.21.11. Also makes `gencfg` write a legacy-friendly `file` list. |
| `-color` | Enable colored output. |

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

- A missing `<world>/saver.json` is created **empty** — no dimension and no file rules. That keeps the list commands usable, and `gencfg <world>` later fills that empty file with the defaults (it refuses to touch a file that already has rules). A backup of an empty rule file produces an archive with nothing in it, and reports success, so watch the archive size.
- Everything named in the rule file must exist in the world: a configured dimension without its `dimensions/<namespace>/<id>` directory, or a `file` entry that is missing, aborts the whole backup instead of being skipped. That is deliberate — otherwise a typo or a stale rule would silently produce an incomplete archive. Remove what you don't have with `del-dms` / `del-file`.
- Missing *region* files are a different matter: a region that was never generated — or a whole missing `region` directory inside an existing dimension — is skipped with a warning, because that is world content rather than something the rule file declares.
- A region listed more than once — by overlapping `range` rules, or by a `range` and a `simple` — is added to the archive once per occurrence.
- Legacy worlds: `gencfg -l <world>` writes a legacy-friendly `file` list (`playerdata` instead of `players`, plus `advancements`).
