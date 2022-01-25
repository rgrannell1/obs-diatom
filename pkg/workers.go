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
	Stats *Stats
	Jobs  chan string
	Count int
}

/*
 * take the channel, read jobs, write metadata to sqlite.
 *
 */
func (work *ExtractWorkers) AnalyseNote(conn *ObsidianDB) <-chan error {
	errChan := make(chan error)

	// writes to errchan
	go func() {
		defer close(errChan)

		// extract information for each file
		for fpath := range work.Jobs {
			work.Stats.Add(COUNT_EXTRACT_NOTE)
			note := NewNote(fpath)

			// walk through markdown note
			walkFailed := false
			for err := range note.Walk(conn) {
				walkFailed = true
				errChan <- err
			}

			// note extraction failed, ignore this note.
			if walkFailed {
				work.Stats.Add(COUNT_FAILED_WALK)
				continue
			}

			// extract data from the note
			done, err := note.ExtractData(conn)

			if err != nil {
				work.Stats.Add(COUNT_FAILED_EXTRACTION)
				errChan <- errors.Wrap(err, "failure extracting data")
				continue
			}

			// if we have analysed this file-hash already; assume the
			// database contains all relevant information for this file
			if done {
				work.Stats.Add(COUNT_NOTE_CACHED)
				continue
			} else {
				work.Stats.Add(COUNT_NOTE_UPDATED)
			}

			// sends errors to errChan
			for err := range note.Write(conn) {
				errChan <- err
			}
		}
	}()

	return errChan
}

/*
 * Extract note information into a SQLite database.
 *
 */
func (work *ExtractWorkers) Start(conn *ObsidianDB, markdownFiles []string) <-chan error {
	var wg sync.WaitGroup
	wg.Add(work.Count)

	results := make(chan error, work.Count)

	// wait to start note extractors
	go func() {
		defer close(results)

		// start workers to process files, and feed erros into a aggregated error channel
		for procId := 0; procId < work.Count; procId++ {
			// start extract worker, forward results
			go func() {
				for err := range work.AnalyseNote(conn) {
					results <- err
				}

				wg.Done()
			}()
		}

		wg.Wait()
	}()

	// write each file to the work channel without blocking
	go func() {
		defer close(work.Jobs)

		for _, fpath := range markdownFiles {
			work.Jobs <- fpath
		}
	}()

	return results
}

// ================================================ //

type GraphWorker struct {
	Stats *Stats
}

/*
 * Start worker to compute and save in & out degree for each
 * file
 *
 */
func (worker *GraphWorker) Start(conn *ObsidianDB) {
	for err := range InDegreeJob(conn) {
		panic(err)
	}

	for err := range OutDegreeJob(conn) {
		panic(err)
	}
}

// ================================================ //

type RemoveWorker struct {
	Stats *Stats
}

/*
 * Start worker to remove deleted files
 *
 */
func (work *RemoveWorker) Start(conn *ObsidianDB) {
	// remove files that do not exist
	for err := range RemoveDeletedFiles(conn) {
		panic(err)
	}
}
