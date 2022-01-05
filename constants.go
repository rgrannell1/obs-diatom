package main

import (
	"database/sql"
	"sync"
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
	db   *sql.DB
	lock *sync.Mutex
}

// CLI Arguments
type DiatomArgs struct {
	dir    string
	dbPath string
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

const Usage = `
Usage:
  diatom (<dpath>)
	diatom (-h | --help)

Description:
  Extract structured data from an Obsidian vault.
`
