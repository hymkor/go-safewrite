CHANGELOG
=========
( [English](./CHANGELOG.md) / Japanese )

v0.4.0
------
Feb 28, 2026

- アプリケーション終了時にまとめてパーミッション復元処理を行うサブパッケージ `perm` を追加
- `io.WriteCloser` のかわりに `File` を定義

v0.3.0
-------
Feb 25, 2026

- パーミッションの復元処理を `Close` から分離し、`RestorePerm` で明示的に行うものとした (#4)
- Info構造体に OverWritten フィールドを追加し、以前、呼び出した時に新規作成したか、更新したかの記録を参照できるようにした。 (#5)

v0.2.0
------
Feb 25, 2026

- バックアップ失敗時に、作業中の一時ファイルの位置を取得できるようにした
  - `BackupError` に一時ファイル名を保持する `Tmp` フィールドを追加
- エラー後のリカバリ処理を共通化するための API を追加
  - `BackupError` / `ReplaceError` に一時ファイル名を取得する `WorkingFile()` メソッドを追加
  - `error` と `WorkingFile()` を持つ `WorkingFileError` インターフェースを定義

v0.1.0
------
Feb 24, 2026

- [試作品](https://github.com/hymkor/bine) (github.com/hymkor/bine/internal/safewrite) よりソースをコピー
