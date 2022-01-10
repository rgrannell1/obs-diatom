package diatom

import (
	_ "github.com/mattn/go-sqlite3"

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

	matches, err := vault.GetNotes("*.md")
	if err != nil {
		return err
	}

	conn, err := NewDB(args.dbPath)
	if err != nil {
		return err
	}

	err = conn.CreateTables()
	if err != nil {
		return err
	}

	defer conn.Close()

	workerCount := 20

	jobs := make(chan string)
	errors := make(chan error)

	// start workers to process files, feed errors into an aggregator channel
	for workIdx := 0; workIdx < workerCount; workIdx++ {
		go func() {
			for err := range ExtractWriteWorker(&conn, jobs) {
				errors <- err
			}
		}()
	}

	// concurrently save file degrees
	go func() {
		for err := range InDegreeJob(&conn) {
			errors <- err
		}
	}()
	go func() {
		for err := range OutDegreeJob(&conn) {
			errors <- err
		}
	}()

	// write each file to a channel read by many workers
	go func() {
		for _, fpath := range matches {
			jobs <- fpath
		}
		close(jobs)
	}()

	for err := range errors {
		return err
	}

	return nil
}

func InDegreeJob(conn *ObsidianDB) <-chan error {
	errors := make(chan error, 0)

	go func() {
		defer close(errors)

		err := conn.AddInDegree()

		if err != nil {
			errors <- err
			return
		}
	}()

	return errors
}

func OutDegreeJob(conn *ObsidianDB) <-chan error {
	errors := make(chan error, 0)

	go func() {
		defer close(errors)

		err := conn.AddOutDegree()

		if err != nil {
			errors <- err
			return
		}
	}()

	return errors
}

// take the channel, read jobs, write metadata to sqlite
func ExtractWriteWorker(conn *ObsidianDB, jobs <-chan string) <-chan error {
	// before exit, decrement the done count.

	errors := make(chan error)

	go func() {
		for fpath := range jobs {
			// extract information
			note := NewNote(fpath)

			// update note in-place
			err := note.Walk(conn)
			if err != nil {
				errors <- err
				continue
			}

			// only extract hash-data where not null
			done, err := note.ExtractData(conn)
			if err != nil {
				errors <- err
				continue
			}

			// if we have analysed this file-hash; assume the database
			// contains all relevant information for this file
			if done {
				continue
			}

			note.Write(conn, errors)
		}
	}()

	return errors
}
