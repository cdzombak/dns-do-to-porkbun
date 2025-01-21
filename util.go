package main

import (
	"fmt"
	"os"
)

func Eprintln(s string) {
	_, _ = fmt.Fprintln(os.Stderr, s)
}
