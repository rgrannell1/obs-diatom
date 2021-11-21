package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gernest/front"
	"github.com/gomarkdown/markdown"
	ast "github.com/gomarkdown/markdown/ast"
)

const obsidianDir = "/home/rg/Drive/Obsidian"
const READ_JOB_COUNT = 3

type yamlInput = map[string]interface{}

// Read from a file
func ReadContent(fpath string) (string, error) {
	body, err := ioutil.ReadFile(fpath)

	if err != nil {
		return "", err
	}

	return string(body), err
}

// Write metadata to Sqlite
func WriteMetadata(fpath string, frontMatter yamlInput, body string) error {
	return nil
}

func GetType(val interface{}) (res string) {
	typeof := reflect.TypeOf(val)
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		res += "*"
	}
	return res + typeof.Name()
}

func ReadMarkdown(body string) string {
	root := markdown.Parse([]byte(body), nil)

	// walk through ast
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		// AZ
		switch node.(type) {
		case *ast.Text:
			leaf := node.AsLeaf()
			fmt.Println(string(leaf.Literal))
		}

		return ast.GoToNext
	})

	// todo
	return ""
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
	frontMatter, body, err := matter.Parse(strings.NewReader(text))
	ReadMarkdown(body)

	if err != nil {
		return err
	}

	return WriteMetadata(fpath, frontMatter, body)
}

// Main function. Read from Obsidian & save as structured data.
func main() {
	matches, err := filepath.Glob(obsidianDir + "/*.md")

	if err != nil {
		panic(err)
	}

	for _, fpath := range matches {
		StoreNoteMetadata(fpath)
	}
}
