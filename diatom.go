package main

import (
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

// Find section bounds in a markdown document, to
// help find the bounds YAML metadata exists within
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

	conn, err := NewDB(args.dbPath)

	if err != nil {
		return err
	}

	defer conn.Close()
	if err != nil {
		return err
	}

	// job, error channels
	workerCount := 20
	jobs := make(chan string)
	errors := make(chan error, workerCount)

	// take the channel, read jobs, write metadata to sqlite
	extractWriteWorker := func(wg *sync.WaitGroup, errors chan<- error) {
		// before exit, decrement the done count.

		for fpath := range jobs {
			// extract information
			note := NewNote(fpath)
			note.Walk(&conn)

			// only extract hash-data where not null
			done, err := note.ExtractData(&conn)
			if err != nil {
				errors <- err
				wg.Done()
				continue
			}

			// if we have analysed this file-hash
			if done {
				wg.Done()
				continue
			}

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
	wg.Add(len(matches))

	// start workers to process files
	for workIdx := 0; workIdx < workerCount; workIdx++ {
		go extractWriteWorker(&wg, errors)
	}

	// write each file to a channel read by many workers
	for _, fpath := range matches {
		jobs <- fpath
	}

	close(jobs)

	// receive errors and panic if received
	select {
	case err := <-errors:
		return err
	default:
	}

	wg.Wait()

	close(errors)

	err = conn.AddInDegree()
	if err != nil {
		return err
	}

	// extract code-blocks

	return nil
}
