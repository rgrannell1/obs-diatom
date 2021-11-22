package main

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
