// cmd/mwb/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "mwb: Mouse Without Borders Linux client")
	os.Exit(0)
}
