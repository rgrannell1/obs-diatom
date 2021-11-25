package main

import (
	"github.com/docopt/docopt-go"
	_ "github.com/mattn/go-sqlite3"
)

// Main function. Read from Obsidian & save as structured data.
func main() {
	opts, err := docopt.ParseDoc(Usage)

	if err != nil {
		panic(err)
	}

	dpath, _ := opts.String("<dpath>")

	args := &DiatomArgs{
		dir: dpath,
	}

	err = Diatom(args)

	if err != nil {
		panic(err)
	}
}
