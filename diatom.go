package main

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"io/ioutil"
	"strings"
)

// Read from a file
func ReadContent(fpath string) (string, error) {
	body, err := ioutil.ReadFile(fpath)

	if err != nil {
		return "", err
	}

	return string(body), err
}

func GetSectionBounds(text string) []int {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	bounds := []int{}

	for idx, line := range lines {
		if strings.HasPrefix(line, "---") {
			bounds = append(bounds, idx)
		}
	}

	return bounds
}

// Main function. Read from Obsidian & save as structured data.
func Diatom(args *DiatomArgs) error {
	vault := ObsidianVault{
		dpath: args.dir,
	}

	matches, err := vault.GetNotes("*.md")

	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", "./diatom.sqlite")

	if err != nil {
		return err
	}

	conn := ObsidianDB{db}
	err = conn.DropTables()

	defer conn.Close()
	if err != nil {
		return err
	}

	// job, error channels
	jobCount := 20
	jobs := make(chan string)
	errors := make(chan error, jobCount)

	// take the channel, read jobs, write metadata to sqlite
	extractWriteWorker := func(wg *sync.WaitGroup, errors chan<- error) {
		// before exit, decrement the done count.

		defer wg.Done()

		for fpath := range jobs {
			// extract information
			note := ObsidianNote{fpath, nil, nil}
			err := note.ExtractData()

			// bail out if extract-data fails
			if err != nil {
				errors <- fmt.Errorf("note.ExtractData() %v: %v", fpath, err)
				return
			}

			note.Write(conn, errors)
			wg.Done()
			// default case, mark job as done
		}
	}

	// distribute among a list of jobs
	var wg sync.WaitGroup
	for jobIdx := 0; jobIdx < jobCount; jobIdx++ {
		go extractWriteWorker(&wg, errors)
	}

	for _, fpath := range matches {
		jobs <- fpath
		wg.Add(1)
	}

	close(jobs)

	wg.Wait()

	// receive errors and panic if received
	select {
	case err := <-errors:
		panic(err)
	default:
	}

	close(errors)

	db.Close()

	return nil
}
