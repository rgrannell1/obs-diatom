package main

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/gernest/front"
)

func (note *ObsidianNote) Read() (string, error) {
	body, err := ioutil.ReadFile(note.fpath)

	if err != nil {
		return "", err
	}

	return string(body), err
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

func (note *ObsidianNote) Write(conn ObsidianDB) error {
	fpath := note.fpath
	bodyData := note.data

	err := conn.CreateTables()
	if err != nil {
		return err
	}

	err = conn.InsertFile(fpath)
	if err != nil {
		return err
	}

	err = conn.InsertTags(bodyData, fpath)
	if err != nil {
		return err
	}

	err = conn.InsertUrl(bodyData, fpath)
	if err != nil {
		return err
	}

	err = conn.InsertWikilinks(bodyData, fpath)
	if err != nil {
		return err
	}

	return nil
}

func (note *ObsidianNote) ExtractData() error {
	text, err := note.Read()

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

	note.frontMatter = frontMatter

	bounds := GetSectionBounds(text)

	// get the suffix of text
	body := text[bounds[1]:]
	extracted := &MarkdownData{
		Title:     FindTitle(body),
		Wikilinks: FindWikilinks(body),
		Tags:      FindTags(body),
		Urls:      FindUrls(body),
	}

	note.data = extracted

	if err != nil {
		return err
	}

	return nil
}
