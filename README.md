# mc-saver

A lightweight Minecraft world backup tool written in Go. It packs a *selected subset* of a world save — region files and core data files — into a dated ZIP archive, based on a JSON rule file.

中文版：[README.zh-CN.md](README.zh-CN.md)

## Features

- **Selective backups** — back up only the chunks you care about instead of the whole world, which keeps archives small and fast.
- **Per-dimension rules** — configure each dimension (`overworld`, `the_nether`, `the_end`, or any custom dimension) independently.
- **Flexible region selection** — select regions by rectangular `range` or by individual `simple` coordinates.
- **Core data support** — include files and folders at the world root, such as `level.dat`, `data`, `datapacks`, and `players`.
- **Dated ZIP output by default** — archives are named `<world>-YYYY-MM-DD.zip` automatically, and you can also pass any output filename directly.
- **Cross-platform** — build for Linux, Windows, and macOS (amd64/arm64) with one script.
- **Simple CLI** — subcommands (`run`, `gencfg`, `help`, `repl`) plus an interactive wizard when no command is given.

## Requirements

- Go 1.26.5+ to build from source (or use the program from a Release).
- A Minecraft Java Edition world directory. The tool verifies that the target path exists and is a directory.

## Build

Build for your current platform:

```bash
go build -o mc-saver .
```

Or cross-compile for all supported platforms with the included script (outputs go to `build/`):

```bash
./build.sh
```

## Usage

### Syntax

```
mc-saver [-c <config file>] [-l] <command> [args...]
```

Flags must be placed before the subcommand; anything after the subcommand is treated as a positional argument.

### Commands

| Command | Defaults | Description |
| --- | --- | --- |
| `run` | `world`, `.` | Back up a world. Positional args: `<world>` and `<output>` (see [Output path](#output-path)). |
| `gencfg` | `save-rule.json` | Write a default config file and exit; refuses to overwrite an existing file. Optional positional arg: `<config file>`. |
| `help` | — | Print the built-in help text and exit. |
| `repl` | — | Run the interactive wizard (same as running with no command). |
| *(no command)* | — | Run the interactive wizard: prompts for the world directory and output path, and offers to generate a config if missing. |

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-c` | `save-rule.json` | Path to the JSON backup rule file. |
| `-l` | off | Back up using the legacy single-folder world layout (`DIM-1`/`DIM1`), see [Legacy mode](#legacy-mode--l). |

### Examples

```bash
# Interactive wizard: enter the world directory and output path
./mc-saver

# Same wizard, explicitly
./mc-saver repl

# Print the built-in help
./mc-saver help

# Back up a specific world; dated archive lands in the current directory
./mc-saver run /path/to/world

# Use a custom rule file
./mc-saver -c rules.json run /path/to/world

# Write the archive into an existing output directory
./mc-saver run /path/to/world /path/to/backups

# Write the archive to a custom output filename
./mc-saver run /path/to/world ~/backups/world-backup.zip

# Generate a default save-rule.json and exit
./mc-saver gencfg

# Generate a default config under a custom name
./mc-saver gencfg rules.json

# Back up using the legacy world layout (-l flag, see below)
./mc-saver -l run /path/to/old_world /path/to/backups
```

### Interactive wizard

With no subcommand (or with `repl`), `mc-saver` walks you through the backup step by step:

1. If the config file (see `-c`, default `save-rule.json`) is missing, it asks whether to generate a default one. Answer `y` (or `Y`) to create it and keep going; the wizard then asks you to confirm (`y`) before continuing. Any other answer exits.
2. Enter the world directory to back up (default `world`).
3. Enter the output path (default `.`, see [Output path](#output-path)).

As soon as the last prompt is answered, the backup starts.

### Output path

The output argument can be:

- an **existing directory** — the archive is written as `<world>-YYYY-MM-DD.zip` inside it; if a file with the same name already exists, a `-1`, `-2`, … suffix is appended automatically so old backups are never overwritten;
- a **file path** (existing or not) — the archive is written directly to that path, so you can choose any name. Missing parent directories are created automatically. If the target file already exists, the same `-1`, `-2`, … numbering is applied instead of overwriting.

Note that a non-existent output path is always treated as a filename, never as a directory. To write into a new directory, create it first (e.g. `mkdir -p /path/to/backups`) or use a path ending in a filename.

## Legacy mode (-l)

Older Minecraft versions use a single-folder layout: overworld content lives directly in the world root, with the Nether and the End stored as `DIM-1` and `DIM1` folders instead of modern `dimensions/<namespace>/<id>/`. Pass the `-l` flag (before the subcommand) to back up using the legacy layout:

```bash
./mc-saver -l run /path/to/old_world /path/to/backups
```

`-l` works with both `run` and the interactive wizard, and must be placed before the subcommand.

In the legacy layout the overworld's data directory is the root-level `data/`, which the dimension walk already covers. Do not list `"data"` in `file` again, or the same files will be packed twice. If you run legacy mode with a config generated by `gencfg`, remove `"data"` from the `file` list first.

## Configuration

The rule file is JSON with two top-level sections, `dimension` and `file`:

```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    },
    "minecraft:the_nether": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    },
    "minecraft:the_end": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
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

Keyed by dimension namespace ID (`<namespace>:<id>`). Each dimension rule can have two selectors:

- `range` — an array of rectangles. Each entry is `{ "from": [x, z], "to": [x, z] }`; every region file whose coordinates fall inside the rectangle (inclusive) is included.
- `simple` — an array of individual region coordinates, e.g. `[x, z]`, for fine-grained selection.

Region coordinates follow Minecraft's region file naming (`r.<x>.<z>.mca`, one region covers 512×512 blocks). If both `range` and `simple` are present for a dimension, both rules apply.

For every selected region of a dimension, the tool collects `r.<x>.<z>.mca` from the `region/`, `entities/`, and `poi/` directories under `dimensions/<namespace>/<id>/`, and also includes all files under that dimension's `data/` directory (skipped silently if it doesn't exist).

Every dimension listed in the config must have a real `dimensions/<namespace>/<id>/` directory, or the backup fails. Nether/End directories usually only appear after a player first enters them, so remove rules for dimensions that don't exist in the save yet.

### `file`

A list of files or folders at the world root to include. Files are added directly; folders are traversed recursively.

Missing region files are reported with a warning and skipped, so sparse dimensions don't abort the backup. Entries in the `file` list are stricter: if a listed file or folder doesn't exist, the backup fails, so make sure every path is actually present (for example, older or freshly generated worlds may not have `datapacks` or `players`).

The `file` list should only contain root-level entries that the dimension rules don't already cover. Do not list `dimensions` — it bypasses `range`/`simple` selection, packs every region file in the tree, and duplicates entries already added by the dimension rules.

## Project layout

```
.
├── main.go          # CLI entry point: flags, validation, ZIP creation
├── parse/           # Config parsing and file collection (writes into the ZIP via callback)
├── record/          # Colored console logging (INFO/WARN/ERROR)
├── build.sh         # Cross-compile script (Linux/Windows/macOS, amd64/arm64)
├── clean.sh         # Developer-only script that resets the repository and force-pushes (dangerous)
├── save-rule.json   # Backup rule file (generate one with gencfg)
└── build/           # Output directory for build.sh binaries
```

## Notes

- The archive keeps the world's directory structure, including the top-level world folder name, so a backup can be dropped back in as a world save directly.
- If you run `mc-saver run .` from inside the world folder, files are stored relative to the current directory (like `zip -r archive.zip .`), so entries have no top-level folder name — extract them directly where the world should live.
- The default config backs up the 3×3 region area from `(-1, -1)` to `(1, 1)` in all three vanilla dimensions — adjust it to match the area you actually play in.
