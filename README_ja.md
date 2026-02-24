go-safewrite
=============
( [English](README.md) / Japanese )

go-safewrite は、ファイルの差し替え・更新を安全に行うための、書き込み用ファイルオープン関数を提供するライブラリです。

```go
package safewrite // import "github.com/hymkor/go-safewrite"

var (
    ErrOverWriteRejected = errors.New("overwrite rejected")
)

func Open(
    name string,
    confirmOverwrite func(*Info) bool,
) (io.WriteCloser, error)
```

Open は io.WriteCloser を返す。返り値の実体は条件により `*os.File` または内部実装型となるが、呼び出し側は `io.WriteCloser` としてのみ扱うことを想定する。

動作
---

- name は更新を行う対象のファイル名
- confirmOverwrite はファイルが既存だった場合に、上書きするかどうかを確認するコールバック関数。これが true を返す場合処理を継続。false の場合 Open は `ErrOverWriteRejected` を返して終了する

`Open` の動作は、既存ファイルの状態に依存する:

- 存在しなかった場合は os.Create と全く同じ動作となり、戻り値の実体は `*os.File`
- 同名のデバイスファイルが存在した場合は、普通に上書きオープンをする。戻り値の実体は `*os.File`
- 同名のプレーンファイル（通常のファイル）が存在した場合は、別名で作成し、Close 時に古いものを今作ったものに差し替えるようにする
  - 古いファイルは、末尾に `~` を付加した名前にリネームしてバックアップする。`~` を付加した名前のファイルが既存の場合は上書きする
  - ただし、同じプロセス内で再度 Open する場合は、初回に作成したバックアップを温存し、以降の保存ではバックアップを更新しない（頻繁な保存操作でバックアップが無意味になることを防ぐため）

Close 時のリネーム処理などが失敗した場合、仕掛り中の一時ファイルは削除せず、そのまま残した上でエラーを返す

Example
-------

```example.go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"

    "github.com/hymkor/go-safewrite"
)

func mains() error {
    prompt := func(info *safewrite.Info) bool {
        sc := bufio.NewScanner(os.Stdin)
        for {
            if info.ReadOnly() {
                fmt.Printf("Overwrite READONLY file %q ? ", info.Name)
            } else {
                fmt.Printf("Overwrite file %q ? ", info.Name)
            }
            if !sc.Scan() {
                return false
            }
            ans := sc.Text()
            if strings.EqualFold(ans, "y") {
                return true
            }
            if strings.EqualFold(ans, "n") {
                return false
            }
        }
    }
    fd, err := safewrite.Open("sample.out", prompt)
    if err != nil {
        return err
    }
    fmt.Fprintln(fd, "sample output.")
    return fd.Close()
}

func main() {
    if err := mains(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())

        var e *safewrite.BackupError
        if errors.As(err, &e) {
            fmt.Fprintln(os.Stderr, "Working file left at:", e.Tmp)
        }

        os.Exit(1)
    }
}
```
