package diatom

import (
	"database/sql"
	"os"
	"path/filepath"
)

// Wikilink data-structure
type Wikilink struct {
	Reference string
	Alias     string
}

// All extracted data from markdown
type MarkdownData struct {
	Title     string
	Wikilinks []*Wikilink
	Tags      []string
	Urls      []string
	Hash      uint32
}

// Obsidian database structure
type ObsidianDB struct {
	Db *sql.DB
}

// CLI Arguments
type DiatomArgs struct {
	Dir    string
	DBPath string
}

// Obsidian note information
type ObsidianNote struct {
	fpath       string
	frontMatter map[string]interface{}
	data        *MarkdownData
}

// Obsidian vault data
type ObsidianVault struct {
	dpath string
}

type File struct {
	id    string
	title string
	hash  string
}

func Usage() string {
	home, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	dbPath := filepath.Join(home, ".diatom.sqlite")

	return `
	Usage:
		diatom (<dpath>) [--dbpath <dbpath>]
		diatom (-h | --help)

	Description:
		Extract structured data from an Obsidian vault into a sqlite database.

	Options:
		--dbpath <dbpath>    the path the diatom sqlite database [default: ` + dbPath + `]
	`
}
