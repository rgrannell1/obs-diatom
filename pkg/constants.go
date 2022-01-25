package diatom

import (
	"database/sql"
	"os"
	"path/filepath"
)

const WORKER_COUNT = 20

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

/*
 * Generate a usage-document
 */
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
  --dbpath <dbpath>       the path the diatom sqlite database [default: ` + dbPath + `]

License:
	The MIT License

	Copyright (c) 2021 Róisín Grannell

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
	associated documentation files (the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
	and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do
	so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or
	substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
	NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
	SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`
}

const COUNT_FAILED_WALK = "count/failed_walk"
const COUNT_FAILED_EXTRACTION = "count/failed_extraction"
const COUNT_EXTRACT_NOTE = "count/extract_note"
const COUNT_NOTE_CACHED = "count/note_cached"
const COUNT_NOTE_UPDATED = "count/note_updated"
