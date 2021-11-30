package main

import (
	"database/sql"
	"path"
	"strings"
)

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
	tx, err := conn.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

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
	ON CONFLICT (id) DO UPDATE SET title = ?, basename = ?, hash = ?
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

func (conn *ObsidianDB) AddDegree() error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		WITH var_in_degree AS (SELECT count(reference) as in_degree, file_id
				FROM file
			LEFT JOIN wikilink ON file.basename = wikilink.reference
				WHERE file_id IS NOT NULL
			GROUP BY file_id
		ORDER BY in_degree DESC)

		INSERT INTO file (file_id, in_degree)
		var_in_degree
		`)

	if err != nil {
		return err
	}

	return tx.Commit()

}
