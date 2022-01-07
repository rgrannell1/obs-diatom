package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/google/gops/agent"
	_ "github.com/mattn/go-sqlite3"
)

// Main function. Read from Obsidian & save as structured data.
func main() {
	opts, err := docopt.ParseDoc(Usage())

	if err != nil {
		panic(err)
	}

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	dpath, _ := opts.String("<dpath>")
	dbpath, _ := opts.String("<dbpath>")

	home, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	dbpath = filepath.Join(home, ".diatom.sqlite")

	args := &DiatomArgs{
		dir:    dpath,
		dbPath: dbpath,
	}

	err = Diatom(args)

	if err != nil {
		panic(err)
	}
}
