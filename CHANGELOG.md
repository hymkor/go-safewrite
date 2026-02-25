CHANGELOG
=========
( English / [Japanese](./CHANGELOG_ja.md) )

v0.3.0
-------
Feb 25, 2026

- Move permission restoration from `Close` to `RestorePerm` (#4)
- Add `Status` field to `Info` to indicate whether the file was created or overwritten before (#5)

v0.2.0
------
Feb 25, 2026

- Add temporary working file path information to BackupError
  - Introduce `Tmp` field on `BackupError`
- Provide a unified way to access leftover working files after errors
  - Add `WorkingFile()` method to `BackupError` and `ReplaceError`
  - Define `WorkingFileError` interface, which exposes `error` and `WorkingFile() string`

v0.1.0
------
Feb 24, 2026

- Copy sources from [prototype](https://github.com/hymkor/bine) (github.com/hymkor/bine/internal/safewrite)
