package main

import (
	"fmt"
	"os"

	"github.com/bitfield/checkepub"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s EPUB_FILE_PATH\n", os.Args[0])
		os.Exit(1)
	}
	result, err := checkepub.Check(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(result)
	if result.Status == checkepub.StatusInvalid {
		os.Exit(1)
	}
}
