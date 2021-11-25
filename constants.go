package main

import "database/sql"

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
}

type ObsidianDB struct {
	db *sql.DB
}

type DiatomArgs struct {
	dir string
}

type ObsidianNote struct {
	fpath       string
	frontMatter map[string]interface{}
	data        *MarkdownData
}

type ObsidianVault struct {
	dpath string
}

const Usage = `
Usage:
  diatom (<dpath>)

Description:
  Foo
`
