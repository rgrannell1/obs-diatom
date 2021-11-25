package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"io/ioutil"
	"strings"
)

// Read from a file
func ReadContent(fpath string) (string, error) {
	body, err := ioutil.ReadFile(fpath)

	if err != nil {
		return "", err
	}

	return string(body), err
}

func GetSectionBounds(text string) []int {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	bounds := []int{}

	for idx, line := range lines {
		if strings.HasPrefix(line, "---") {
			bounds = append(bounds, idx)
		}
	}

	return bounds
}

// Main function. Read from Obsidian & save as structured data.
func Diatom(args *DiatomArgs) error {
	vault := ObsidianVault{
		dpath: args.dir,
	}

	matches, err := vault.GetNotes("*.md")

	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", "./diatom.sqlite")
	if err != nil {
		return err
	}

	conn := ObsidianDB{db}
	err = conn.DropTables()

	defer conn.Close()
	if err != nil {
		return err
	}

	for _, fpath := range matches {
		note := ObsidianNote{fpath, nil, nil}
		err := note.ExtractData()

		if err != nil {
			return err
		}

		err = note.Write(conn)

		if err != nil {
			return err
		}
	}

	return nil
}
