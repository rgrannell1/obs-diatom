package diatom

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gernest/front"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

/*
 * Construct an Obsidian note representation
 */
func NewNote(fpath string) ObsidianNote {
	return ObsidianNote{fpath, nil, nil}
}

/*
 * Read an Obsidian note from a file
 */
func (note *ObsidianNote) Read() (string, error) {
	body, err := ioutil.ReadFile(note.fpath)

	if err != nil {
		return "", fmt.Errorf("note.Read() %v: %v", note.fpath, err)
	}

	return string(body), err
}

/*
 * Find the title in the target Obsidian file
 */
func (note *ObsidianNote) FindTitle() string {
	title := regexp.MustCompile("-")
	pair := title.Split(strings.Replace(note.fpath, ".md", "", 1), -1)

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
	wg.Add(5)

	// insert file
	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertFile(fpath, bodyData.Title, fmt.Sprint(bodyData.Hash))
		if err != nil {
			panic(err)
		}
	}(errors)

	// insert tags
	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertTags(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	// insert url
	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertUrl(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	// insert wikilink
	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertWikilinks(bodyData, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	// insert frontmatter
	go func(errors chan<- error) {
		defer wg.Done()

		err := conn.InsertFrontmatter(note.frontMatter, fpath)
		if err != nil {
			panic(err)
		}
	}(errors)

	wg.Wait()

	if err != nil {
		panic(err)
	}
}

/*
 * Get the hash currently recorded for a file
 */
func (note *ObsidianNote) GetStoredHash(conn *ObsidianDB) (string, error) {
	return conn.GetFileHash(note.fpath)
}

/*
 * Extract information about a note
 */
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
		Title:     note.FindTitle(),
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

/*
 * Read and parse note content as markdown
 */
func (note *ObsidianNote) Parse() (ast.Node, error) {
	content, err := os.ReadFile(note.fpath)
	if err != nil {
		return nil, err
	}

	return parser.New().Parse(content), nil
}

/*
 * Walk through note and collect interesting information
 */
func (note *ObsidianNote) Walk(conn *ObsidianDB) error {
	doc, err := note.Parse()

	if err != nil {
		return err
	}

	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}

	// walk through markdown tree and store interesting information
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		switch node.(type) {
		case *ast.CodeBlock:
			leaf := node.AsLeaf()

			info := string(node.(*ast.CodeBlock).Info)

			if len(info) > 0 && info[0] == '!' {
				// -- a special code-block containing application-readable data

				yaml := string(leaf.Literal)
				err = conn.InsertMetadata(tx, note.fpath, info, yaml)

				if err != nil {
					panic(err)
				}
			}
		}

		return ast.GoToNext
	})

	tx.Commit()

	return nil
}
