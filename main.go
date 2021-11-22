package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gernest/front"
)

const DB_PATH = "./diatom.sqlite"
const obsidianDir = "/home/rg/Drive/Obsidian"

type yamlInput = map[string]interface{}

// Read from a file
func ReadContent(fpath string) (string, error) {
	body, err := ioutil.ReadFile(fpath)

	if err != nil {
		return "", err
	}

	return string(body), err
}

func CreateTables(db *sql.DB) error {
	// create a file table
	_, err := db.Query(`CREATE TABLE IF NOT EXISTS file (
		id    TEXT NOT NULL PRIMARY KEY,
		title TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS tag (
		tag TEXT      NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS url (
		url      TEXT NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS wikilink (
		reference TEXT NOT NULL,
		alias    TEXT,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	return nil
}

func InsertFile(db *sql.DB, fpath string) error {
	// save file to table
	_, err := db.Exec("INSERT INTO file (id, title) VALUES (?, ?)", fpath, fpath)
	return err
}

func InsertTags(db *sql.DB, bodyData *MarkdownData, fpath string) error {
	// save tag to table
	for _, tag := range bodyData.Tags {
		_, err := db.Exec("INSERT INTO tag (tag, file_id) VALUES (?, ?)", tag, fpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertUrl(db *sql.DB, bodyData *MarkdownData, fpath string) error {
	// save url to table
	for _, url := range bodyData.Urls {
		_, err := db.Exec("INSERT INTO url (url, file_id) VALUES (?, ?)", url, fpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertWikilinks(db *sql.DB, bodyData *MarkdownData, fpath string) error {
	// save wikilinks to table
	for _, wikilink := range bodyData.Wikilinks {
		_, err := db.Exec("INSERT INTO wikilink (reference, alias, file_id) VALUES (?, ?, ?)", wikilink.Reference, wikilink.Alias, fpath)

		if err != nil {
			return err
		}
	}

	return nil
}

// Write metadata to sqlite
func WriteMetadata(fpath string, frontMatter yamlInput, bodyData *MarkdownData) error {
	db, err := sql.Open("sqlite3", DB_PATH)

	if err != nil {
		return err
	}

	defer db.Close()

	// create tables initially
	err = CreateTables(db)
	if err != nil {
		return err
	}

	err = InsertFile(db, fpath)
	if err != nil {
		return err
	}

	err = InsertTags(db, bodyData, fpath)
	if err != nil {
		return err
	}

	err = InsertUrl(db, bodyData, fpath)
	if err != nil {
		return err
	}

	err = InsertWikilinks(db, bodyData, fpath)
	if err != nil {
		return err
	}

	return nil
}

func FindTitle(body string) string {
	return ""
}

// Find WikiLinks
func FindWikilinks(body string) []*Wikilink {
	wikilinkPattern := regexp.MustCompile(`\[{2}[^\[]+\]{2}`)

	linkText := wikilinkPattern.FindAllString(body, -1)
	wikilinks := make([]*Wikilink, 0)

	for _, link := range linkText {
		text := link[2 : len(link)-2]
		parts := strings.Split(text, "|")

		ref := parts[0]
		alias := ""

		if len(parts) > 1 {
			alias = parts[1]
		}

		wikilinks = append(wikilinks, &Wikilink{
			Reference: ref,
			Alias:     alias,
		})
	}

	return wikilinks
}

// Find tags
func FindTags(body string) []string {
	tagPattern := regexp.MustCompile(`\#[a-zA-Z_]+`)

	return tagPattern.FindAllString(body, -1)
}

// Find URLs
func FindUrls(body string) []string {
	return []string{}
}

// Parse markdown as partially-structured data. Read:
//
// Wikilinks
// Tags
// External URLs
// Title
//
// into a structured format.
func ReadMarkdown(body string) *MarkdownData {
	title := FindTitle(body)
	wikilinks := FindWikilinks(body)
	tags := FindTags(body)
	urls := FindUrls(body)

	return &MarkdownData{
		Title:     title,
		Wikilinks: wikilinks,
		Tags:      tags,
		Urls:      urls,
	}
}

// Store note metadata
func StoreNoteMetadata(fpath string) error {
	text, err := ReadContent(fpath)

	if err != nil {
		return err
	}

	matter := front.NewMatter()
	matter.Handle("---", front.YAMLHandler)

	// read frontmatter and body
	frontMatter, _, err := matter.Parse(strings.NewReader(text))

	if err != nil {
		return err
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	bounds := []int{}

	for idx, line := range lines {
		if strings.HasPrefix(line, "---") {
			bounds = append(bounds, idx)
		}
	}

	// get the suffix of text
	body := text[bounds[1]:]
	extracted := ReadMarkdown(body)

	if err != nil {
		return err
	}

	return WriteMetadata(fpath, frontMatter, extracted)
}

// Main function. Read from Obsidian & save as structured data.
func main() {
	matches, err := filepath.Glob(obsidianDir + "/*.md")

	if err != nil {
		panic(err)
	}

	for _, fpath := range matches {
		err = StoreNoteMetadata(fpath)

		if err != nil {
			panic(err)
		}
	}
}
