package diatom

import (
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

/*
 * Main function. Read from Obsidian & save as structured data.
 *
 */
func Diatom(args *DiatomArgs) error {
	conn, err := NewDB(args.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.CreateTables(); err != nil {
		return errors.Wrap(err, "failure creating tables")
	}

	stats := Stats{
		Data: map[string]int{},
		Lock: &sync.Mutex{},
	}

	// extract information for each note into the database
	extractors := ExtractWorkers{
		Stats: &stats,
		Count: WORKER_COUNT,
		Jobs:  make(chan string, 0),
	}

	vault := ObsidianVault{dpath: args.Dir}
	mdFiles, err := vault.GetNotes()
	if err != nil {
		return err
	}

	for err := range extractors.Start(&conn, mdFiles) {
		panic(err)
	}

	graphers := GraphWorker{
		Stats: &stats,
	}
	graphers.Start(&conn)

	removers := RemoveWorker{
		Stats: &stats,
	}
	removers.Start(&conn)

	return nil
}

/*
 * Insert in-degrees into database
 *
 */
func InDegreeJob(conn *ObsidianDB) <-chan error {
	errorsChan := make(chan error, 0)

	go func() {
		defer close(errorsChan)

		if err := conn.AddInDegree(); err != nil {
			errorsChan <- errors.Wrap(err, "failure setting in-degree")
			return
		}
	}()

	return errorsChan
}

/*
 * Insert out-degrees into database
 */
func OutDegreeJob(conn *ObsidianDB) <-chan error {
	errorsChan := make(chan error, 0)

	go func() {
		defer close(errorsChan)

		if err := conn.AddOutDegree(); err != nil {
			errorsChan <- errors.Wrap(err, "failure setting out-degree")
			return
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
