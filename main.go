package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/google/gops/agent"
	_ "github.com/mattn/go-sqlite3"
)

// Main function. Read from Obsidian & save as structured data.
func main() {
	opts, err := docopt.ParseDoc(Usage)

	if err != nil {
		panic(err)
	}

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	dpath, _ := opts.String("<dpath>")

	args := &DiatomArgs{
		dir:    dpath,
		dbPath: "./diatom.sqlite",
	}

	err = Diatom(args)

	if err != nil {
		panic(err)
	}

}
