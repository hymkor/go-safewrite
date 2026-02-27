package perm

import (
	"github.com/hymkor/go-safewrite"
)

var seen = map[string]safewrite.File{}

func Track(f safewrite.File) {
	name := f.Name()
	if _, ok := seen[name]; !ok {
		seen[name] = f
	}
}

func RestoreAll() error {
	for _, f := range seen {
		if err := safewrite.RestorePerm(f); err != nil {
			return err
		}
	}
	for i := range seen {
		delete(seen, i)
	}
	return nil
}
