# mc-saver

A Minecraft world backup tool written in Go. Based on a JSON rule file, it packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

中文版：[README.zh-CN.md](README.zh-CN.md)

## Features

- **Selective backups** — back up only dimensions and chunks that you care about instead of the whole world, so archives are smaller and faster.
- **Per-dimension configuration** — selectively configure and back up the overworld (`overworld`), the Nether (`the_nether`), the End (`the_end`), and any custom dimension.
- **Flexible region selection** — use rectangular `range` rules, or `simple` rules to specify individual region coordinates.

## Download

Download the corresponding binary executable from the [Releases](https://github.com/fovlin/mc-saver/releases) page.

## Example usage

```bash
./mc-saver
# Starts a wizard that guides you through the backup; equivalent to the repl subcommand

./mc-saver gencfg
# Generates a default config file - save-rule.json

./mc-saver run
# Starts a backup with default values, which are:
# config file - save-rule.json
# world directory - world
# output - world-$time.zip

./mc-saver -c config.json run level level.zip
# Starts a backup with config file config.json, world directory level, and output file level.zip

./mc-saver -l run
# Starts a backup with default values in legacy mode; worlds from before 1.21.11 should use this mode:
```

### Syntax

```
mc-saver [-c <config file>] [-l] <command> [args...]
```

Flags must be placed before the subcommand; anything after the subcommand is treated as a positional argument.

### Commands

| Command | Defaults | Description |
| --- | --- | --- |
| `run` | `world`, `.` | Back up a world. Positional args: `<world>` and `<output>` (see [Output path](#output-path)). |
| `gencfg` | `save-rule.json` | Generate a default config file and exit; throw error if it already exists. Optional positional arg: `<config file>`. |
| `help` | — | Print the built-in help text and exit. |
| `repl` | — | Run the interactive wizard (same as running with no command). |
| *(no command)* | — | Run the interactive wizard: prompts for the world directory and output path in turn, and asks whether to generate a config file if one is missing. |

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-c` | `save-rule.json` | Path to the JSON backup rule file. |
| `-l` | off | Back up using the legacy single-folder layout (`DIM-1`/`DIM1`), see [Legacy layout mode](#legacy-layout-mode-l). |

## Configuration

The rule file is JSON with two top-level fields: `dimension` and `file`.

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

Keyed by dimension namespace ID (`<namespace>:<id>`). Each dimension rule supports two selection methods:

- `range` — an array of rectangles. Each entry is `{ "from": [x, z], "to": [x, z] }`; every region file whose coordinates fall inside the rectangle (inclusive) is backed up.

```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ]
    }
  }
}
```

- `simple` — an array of individual region coordinates, e.g. `[x, z]`, for precisely specifying individual regions.

```json
{
  "dimension": {
    "minecraft:overworld": {
      "simple": [
        [3, 3]
      ]
    }
  }
}
```

Mix `range` and `simple` rules.

```json
{
  "dimension": {
    "minecraft:overworld": {
      "range": [
        { "from": [-1, -1], "to": [1, 1] }
      ],
      "simple": [
        [3, 3]
      ]
    }
  }
}
```

Region coordinates follow Minecraft's region file naming (`r.<x>.<z>.mca`, one region covers 512×512 blocks).

### `file`

A list of files or folders at the world root to include. Files are added directly; folders are traversed recursively.

## Notes

- When using legacy world mode, manually edit the `file` field in the config file: remove `data` and change `players` to `playerdata`.
