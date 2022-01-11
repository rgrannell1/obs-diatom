package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/google/gops/agent"
	_ "github.com/mattn/go-sqlite3"

	diatom "github.com/rgrannell1/diatom/pkg"
)

// Main function. Read from Obsidian & save as structured data.
func main() {
	opts, err := docopt.ParseDoc(diatom.Usage())

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
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	dbpath = filepath.Join(home, ".diatom.sqlite")

	args := &diatom.DiatomArgs{
		Dir:    dpath,
		DBPath: dbpath,
	}

	err = diatom.Diatom(args)

	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
