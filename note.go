package main

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	"github.com/gernest/front"
)

func (note *ObsidianNote) Read() (string, error) {
	body, err := ioutil.ReadFile(note.fpath)

	if err != nil {
		return "", fmt.Errorf("note.Read() %v: %v", note.fpath, err)
	}

	return string(body), err
}

func FindTitle(fpath string) string {
	title := regexp.MustCompile("-")
	pair := title.Split(strings.Replace(fpath, ".md", "", 1), -1)

	return pair[len(pair)-1]
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

func HashContent(text string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(text))
	return hash.Sum32()
}

func (note *ObsidianNote) Write(conn ObsidianDB, errors chan<- error) {
	fpath := note.fpath
	bodyData := note.data

	err := conn.CreateTables()
	if err != nil {
		errors <- fmt.Errorf("note.CreateTables() %v: %v", note.fpath, err)
		return
	}

	if bodyData == nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertFile(fpath, bodyData.Title, fmt.Sprint(bodyData.Hash))
		if err != nil {
			panic(err)
		}
	}(errors)

	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertTags(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertUrl(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertWikilinks(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	wg.Wait()
}

func (note *ObsidianNote) GetStoredHash(conn *ObsidianDB) (string, error) {
	return conn.GetFile(note.fpath)
}

func (note *ObsidianNote) ExtractData(conn *ObsidianDB) (bool, error) {
	text, err := note.Read()

	if err != nil {
		return false, err
	}

	hash, err := note.GetStoredHash(conn)
	if err != nil {
		return false, err
	}

	currHash := HashContent(text)
	if hash == fmt.Sprint(currHash) {
		return true, nil
	}

	matter := front.NewMatter()
	matter.Handle("---", front.YAMLHandler)

	// read frontmatter and body
	frontMatter, _, err := matter.Parse(strings.NewReader(text))

	if err != nil {
		note.frontMatter = map[string]interface{}{}
		return false, nil
	}

	note.frontMatter = frontMatter

	bounds := GetSectionBounds(text)
	body := ""

	if len(bounds) > 0 {
		// get the suffix of text
		body = text[bounds[1]:]
	}

	tags := FindTags(body)

	extracted := &MarkdownData{
		Title:     FindTitle(note.fpath),
		Wikilinks: FindWikilinks(body),
		Tags:      tags,
		Urls:      FindUrls(body),
		Hash:      currHash,
	}

	note.data = extracted

	if err != nil {
		return false, err
	}

	return false, nil
}
