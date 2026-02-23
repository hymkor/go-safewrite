//go:build example

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hymkor/go-safewrite"
)

func mains() error {
	fname := "sample.out"
	prompt := func() bool {
		sc := bufio.NewScanner(os.Stdin)
		for {
			fmt.Printf("Overwrite %q ? ", fname)
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
	fd, err := safewrite.Open(fname, prompt)
	if err != nil {
		return err
	}
	fmt.Fprintln(fd, "sample output.")
	return fd.Close()
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
