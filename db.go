package main

import (
	"database/sql"
	"path"
	"strings"
	"sync"
)

/*
 * Construct a database
 */
func NewDB(fpath string) (ObsidianDB, error) {
	db, err := sql.Open("sqlite3", fpath)

	if err != nil {
		return ObsidianDB{}, err
	}

	return ObsidianDB{db, &sync.Mutex{}}, nil
}

// Close a database connection
func (conn *ObsidianDB) Close() error {
	return conn.db.Close()
}

// Create diatom tables in sqlite
func (conn *ObsidianDB) CreateTables() error {
	// create a file table
	tx, err := conn.db.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS file (
		id         TEXT NOT NULL,
		basename   TEXT NOT NULL,
		title      TEXT NOT NULL,
		hash       TEXT NOT NULL,
		in_degree  INTEGER,
		out_degree INTEGER,

		PRIMARY KEY(id)
	)`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS tag (
		tag TEXT      NOT NULL,
		file_id  TEXT NOT NULL,

		PRIMARY KEY(tag, file_id)
	)`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS url (
		url      TEXT NOT NULL,
		file_id  TEXT NOT NULL,

		PRIMARY KEY(url, file_id)
	)`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS wikilink (
		reference TEXT NOT NULL,
		alias    TEXT,
		file_id  TEXT NOT NULL,

		PRIMARY KEY(reference, alias, file_id)
	)`)

	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (conn *ObsidianDB) GetFile(fpath string) (string, error) {
	file := File{fpath, "", ""}

	row := conn.db.QueryRow(`SELECT * FROM file WHERE file.id = ?`, fpath)
	switch err := row.Scan(&file.id, &file.title, &file.hash); err {
	case sql.ErrNoRows:
		return "", nil
	case nil:
		return file.hash, nil
	}

	return "", nil
}

func (conn *ObsidianDB) InsertFile(fpath, title, hash string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	basename := path.Base(fpath)
	ext := path.Ext(basename)

	idx := strings.LastIndex(basename, ext)
	basename = basename[:idx]

	// save file to table
	_, err = tx.Exec(`
	INSERT OR IGNORE INTO file (id, basename, title, hash) VALUES (?, ?, ?, ?)
	ON CONFLICT (id)
	DO UPDATE SET title = ?, basename = ?, hash = ?
	`, fpath, basename, title, hash, title, basename, hash)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (conn *ObsidianDB) InsertTags(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, tag := range bodyData.Tags {
		_, err := tx.Exec(`
		INSERT OR IGNORE INTO tag (tag, file_id) VALUES (?, ?)
		`, tag, fpath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (conn *ObsidianDB) InsertUrl(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range bodyData.Urls {
		_, err := tx.Exec(`
		INSERT OR IGNORE INTO url (url, file_id) VALUES (?, ?)
		`, url, fpath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (conn *ObsidianDB) InsertWikilinks(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, wikilink := range bodyData.Wikilinks {
		_, err := tx.Exec(`
		INSERT OR IGNORE INTO wikilink (reference, alias, file_id) VALUES (?, ?, ?)
		`, wikilink.Reference, wikilink.Alias, fpath)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (conn *ObsidianDB) GetInDegree() (*sql.Rows, error) {
	return conn.db.Query(`
		SELECT count(reference) as in_degree, id
				FROM file
			LEFT JOIN wikilink ON file.basename = wikilink.reference
				WHERE id IS NOT NULL
			GROUP BY id
		ORDER BY in_degree DESC`)
}

func (conn *ObsidianDB) GetOutDegree() (*sql.Rows, error) {
	return conn.db.Query(``)
}

/*
 * Insert a degree row into the database
 */
func (conn *ObsidianDB) InsertInDegree(tx *sql.Tx, rows *sql.Rows) error {
	file := struct {
		in_degree int
		id        string
	}{}

	err := rows.Scan(&file.in_degree, &file.id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	UPDATE file
	SET in_degree = ?
	WHERE id = ?
	`, file.in_degree, file.id)

	return nil
}

// Add the file degree
func (conn *ObsidianDB) AddInDegree() error {
	rows, err := conn.GetInDegree()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	tx, _ := conn.db.Begin()

	for rows.Next() {
		if err = conn.InsertInDegree(tx, rows); err != nil {
			return err
		}
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	return nil
}
