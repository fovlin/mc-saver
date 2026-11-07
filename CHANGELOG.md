# Changelog

All notable changes to mc-saver are documented here. The project follows
[Semantic Versioning](https://semver.org/).

## 2.0.1 - 2026-11-02

### Fixed

- Tab completion of dimension ids now honours `-l`. In the legacy layout the
  three vanilla ids are always offered, and a world without a `dimensions`
  directory no longer turns completion off.

## 2.0.0 - 2026-11-02

### Breaking changes

- Renamed the config command families to shorter, shell-like names. The old
  names are removed, not aliased:
  - `list`, `list-config`, `list-dms`, `list-range`, `list-simple`, `list-file`
    are now `ls`, `ls-config`, `ls-dms`, `ls-range`, `ls-simple`, `ls-file`.
  - `del-dms`, `del-range`, `del-simple`, `del-file` are now `rm-dms`,
    `rm-range`, `rm-simple`, `rm-file`.
  - `mod-dms`, `mod-range`, `mod-simple`, `mod-file` are now `set-dms`,
    `set-range`, `set-simple`, `set-file`.
- Command arguments are unchanged; only the command names changed.

### Added

- Tab completion in REPL mode for command names and dimension ids.
- `@`, `:` and `=` literal prefixes for integer arguments, so negative
  coordinates can be written on the command line as `@-3`. The prefix is
  stripped before parsing; in the REPL it is optional.

### Internal

- Commands are registered once in `cmdInfo`; `cmdMap` is derived from it.
- The command table records the minimum argument count and whether a command
  takes a dimension as its second argument.

## Earlier releases

See the [tags](https://github.com/fovlin/mc-saver/tags) for releases before
2.0.0.
