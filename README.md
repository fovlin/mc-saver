# mc-saver

A lightweight Minecraft world backup tool written in Go. It packs a *selected subset* of a world save — region files and core data files — into a dated ZIP archive, based on a JSON rule file.

中文版：[README.zh-CN.md](README.zh-CN.md)

## Features

- **Selective backups** — back up only the chunks you care about instead of the whole world, which keeps archives small and fast.
- **Per-dimension rules** — configure each dimension (`overworld`, `the_nether`, `the_end`, or any custom dimension) independently.
- **Flexible region selection** — select regions by rectangular `range` or by individual `simple` coordinates.
- **Core data support** — include files and folders at the world root, such as `level.dat`, `data`, `datapacks`, and `players`.
- **Dated ZIP output** — archives are named `<world>-YYYY-MM-DD.zip` automatically.
- **Cross-platform** — build for Linux, Windows, and macOS (amd64/arm64) with one script.
- **Simple CLI** — just a config file, a world directory, and a run.

## Requirements

- Go 1.26.5+ to build from source (or use the prebuilt binaries in `build/`).
- A Minecraft Java Edition world directory. The tool verifies that the directory exists and contains `level.dat`.

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

```
# Syntax: -c specifies the config file, -t the world directory, -o the output directory
mc-saver [-c <config file>] [-t <world directory>] [-o <output directory>]

# gencfg: write a default config file and exit (place -c before gencfg to set the output path)
mc-saver [-c <config file>] gencfg
```

| Flag | Default | Description |
| --- | --- | --- |
| `-c` | `save-rule.json` | Path to the JSON backup rule file. |
| `-t` | `world` | Path to the world directory to back up. |
| `-o` | `.` | Directory where the ZIP archive is written. |

Examples:

```bash
# Back up ./world using ./save-rule.json
./mc-saver

# Back up a specific world
./mc-saver -t /path/to/world

# Use a custom rule file
./mc-saver -c rules.json -t /path/to/world

# Write the archive to a custom output directory
./mc-saver -t /path/to/world -o /path/to/backups

# Generate a default save-rule.json and exit
./mc-saver gencfg
```

The result is a ZIP archive named after the world and the current date — e.g. `world-2026-08-13.zip` — written to the output directory (`.` by default).

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

For every selected region of a dimension, the tool collects `r.<x>.<z>.mca` from the `region/`, `entities/`, and `poi/` directories under `dimensions/<namespace>/<id>/`, and also includes all files under that dimension's `data/` directory.

### `file`

A list of files or folders at the world root to include. Files are added directly; folders are traversed recursively.

Files that don't exist are skipped instead of failing the backup: missing region files are reported with a warning, while missing entries in the `file` list are skipped silently.

## Project layout

```
.
├── main.go          # CLI entry point: flags, validation, ZIP creation
├── parse/           # Config parsing and file collection (writes into the ZIP via callback)
├── record/          # Colored console logging (INFO/WARN/ERROR)
├── build.sh         # Cross-compile script (Linux/Windows/macOS, amd64/arm64)
├── save-rule.json   # Example backup rule file
└── build/           # Output directory for build.sh binaries
```

## Notes

- The archive keeps the world's directory structure, including the top-level world folder name, so a backup can be dropped back in as a world save directly.
- The default config backs up the 3×3 region area from `(-1, -1)` to `(1, 1)` in all three vanilla dimensions — adjust it to match the area you actually play in.