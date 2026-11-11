# Changelog

All notable changes to mc-saver are documented here. The project follows
[Semantic Versioning](https://semver.org/).

## 2.0.4 - 2026-11-11

### Fixed

- Dimension completion now reads the selected world instead of the command name.
- Pressing Tab on a whitespace-only line no longer panics.
- Pressing Tab after a space now completes the next argument instead of
  appending the previous command again.
- Completion in the middle of a word no longer repeats the text after the
  cursor.
- Restored the completion regression tests.

## 2.0.3 - 2026-11-09

### Fixed

- Completing a dimension id with a single match did nothing: the candidate was
  looked up in the command table, which never holds an id.
- Completion rewrote the whole line and moved the cursor to its end. It now
  replaces only the word under the cursor and keeps the rest, spacing included.

### Changed

- The candidate lists are loaded on every key press instead of once when the
  prompt starts, so dimensions added during the session can be completed right
  away.

## 2.0.2 - 2026-11-08

### Fixed

- Tab completion no longer drops the text that follows the cursor.
- Completing a half typed command line no longer panics.

### Changed

- The completer is a struct whose candidate lists are built once, instead of a
  callback rebuilt on every keystroke.

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
