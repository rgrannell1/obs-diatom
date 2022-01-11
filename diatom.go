package main

import (
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"strings"
)

/*
 * Find section bounds in a markdown document, to
 * help find the bounds YAML metadata exists within
 */
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
	vault := ObsidianVault{dpath: args.dir}

	matches, err := vault.GetNotes("**/*.md")
	if err != nil {
		return err
	}

	conn, err := NewDB(args.dbPath)
	if err != nil {
		return err
	}

	err = conn.CreateTables()
	if err != nil {
		return errors.Wrap(err, "failure creating tables")
	}

	defer conn.Close()

	workerCount := 20

	jobs := make(chan string)
	errorsChan := make(chan error, 1024)

	var wg sync.WaitGroup
	wg.Add((workerCount) + 3) // two more for other queries

	// start workers to process files, feed errorsChan into an aggregator channel
	for workIdx := 0; workIdx < workerCount; workIdx++ {
		go func() {
			for err := range ExtractWriteWorker(&conn, jobs) {
				errorsChan <- err
			}

			wg.Done()
		}()
	}

	// write each file to a channel read by many workers
	go func() {
		for _, fpath := range matches {
			jobs <- fpath
		}
		close(jobs)
	}()

	go func() {
		for err := range InDegreeJob(&conn) {
			errorsChan <- err
		}

		wg.Done()
	}()

	go func() {
		for err := range OutDegreeJob(&conn) {
			errorsChan <- err
		}

		wg.Done()
	}()

	go func() {
		for err := range RemoveDeletedFiles(&conn) {
			errorsChan <- err
		}

		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		return err
	}

	return nil
}

func InDegreeJob(conn *ObsidianDB) <-chan error {
	errorsChan := make(chan error, 0)

	go func() {
		defer close(errorsChan)

		err := conn.AddInDegree()

		if err != nil {
			errorsChan <- errors.Wrap(err, "failure setting in-degree")
			return
		}
	}()

	return errorsChan
}

func OutDegreeJob(conn *ObsidianDB) <-chan error {
	errorsChan := make(chan error, 0)

	go func() {
		defer close(errorsChan)

		err := conn.AddOutDegree()

		if err != nil {
			errorsChan <- errors.Wrap(err, "failure setting out-degree")
			return
		}
	}()

	return errorsChan
}

/*
 * take the channel, read jobs, write metadata to sqlite
 */
func ExtractWriteWorker(conn *ObsidianDB, jobs <-chan string) <-chan error {
	// before exit, decrement the done count.

	errorsChan := make(chan error)

	go func() {
		defer close(errorsChan)

		for fpath := range jobs {
			// extract information
			note := NewNote(fpath)

			// update note in-place
			err := note.Walk(conn)
			if err != nil {
				errorsChan <- errors.Wrap(err, "failure walking through note markdown")
				continue
			}

			// only extract hash-data where not null
			done, err := note.ExtractData(conn)
			if err != nil {
				errorsChan <- errors.Wrap(err, "failure extracting data")
				continue
			}

			// if we have analysed this file-hash; assume the database
			// contains all relevant information for this file
			if done {
				continue
			}

			note.Write(conn, errorsChan)
		}
	}()

	return errorsChan
}

/*
 * Remove database entries for files that no longer exist
 */
func RemoveDeletedFiles(conn *ObsidianDB) <-chan error {
	errorsChan := make(chan error)

	go func() {
		defer close(errorsChan)

		fpaths, err := conn.GetFileIds()
		if err != nil {
			errorsChan <- errors.Wrap(err, "failure checking if note existed")
			return
		}

		for _, fpath := range fpaths {
			note := NewNote(fpath)

			exists, err := note.Exists()
			if err != nil {
				errorsChan <- errors.Wrap(err, "failure checking if note existed")
				continue
			}

			if !exists {
				err := note.Delete(conn)
				if err != nil {
					errorsChan <- errors.Wrap(err, "failure deleting non-existing note from database")
					continue
				}
			}
		}

	}()

	return errorsChan
}
