package diatom

import (
	"sync"

	"github.com/pkg/errors"
)

/*
 * Extract workers manager definition
 *
 */
type ExtractWorkers struct {
	Jobs  chan string
	Count int
}

/*
 * Extract note information into a SQLite database.
 *
 */
func (work *ExtractWorkers) Start(conn *ObsidianDB, markdownFiles []string) <-chan error {
	var wg sync.WaitGroup
	wg.Add(work.Count)

	results := make(chan error, work.Count)

	go func() {
		// start workers to process files, feed errorsChan into an aggregator channel
		for procId := 0; procId < work.Count; procId++ {
			// start extract worker, forward results
			go func() {
				for err := range ExtractNoteData(conn, work.Jobs) {
					results <- err
				}

				wg.Done()
			}()
		}

		wg.Wait()
		close(results)
	}()

	go func() {
		// write each file to a channel
		for _, fpath := range markdownFiles {
			work.Jobs <- fpath
		}
		close(work.Jobs)
	}()

	return results
}

/*
 * take the channel, read jobs, write metadata to sqlite.
 *
 */
func ExtractNoteData(conn *ObsidianDB, fpaths <-chan string) <-chan error {
	errChan := make(chan error)

	// writes to errchan
	go func() {
		defer close(errChan)

		// extract information for each file
		for fpath := range fpaths {
			note := NewNote(fpath)

			// walk through markdown note
			walkFailed := false
			for err := range note.Walk(conn) {
				walkFailed = true
				errChan <- err
			}

			// note extraction failed, ignore this note.
			if walkFailed {
				continue
			}

			// extract data from the note
			done, err := note.ExtractData(conn)

			if err != nil {
				errChan <- errors.Wrap(err, "failure extracting data")
				continue
			}

			// if we have analysed this file-hash already; assume the
			// database contains all relevant information for this file
			if done {
				continue
			}

			// sends errors to errChan
			note.Write(conn, errChan)
		}
	}()

	return errChan
}

// ================================================ //

type GraphWorker struct{}

/*
 * Start worker to compute and save in & out degree for each
 * file
 *
 */
func (worker *GraphWorker) Start(conn *ObsidianDB) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		// write the in-degree for the files present

		go func() {
			for err := range InDegreeJob(conn) {
				panic(err)
			}

			wg.Done()
		}()

		// write the out-degree for the files present
		go func() {
			for err := range OutDegreeJob(conn) {
				panic(err)
			}

			wg.Done()
		}()
	}()

	wg.Wait()
}

// ================================================ //

type RemoveWorker struct{}

/*
 * Start worker to remove deleted files
 */
func (work *RemoveWorker) Start(conn *ObsidianDB) {
	// remove files that do not exist
	for err := range RemoveDeletedFiles(conn) {
		panic(err)
	}
}
