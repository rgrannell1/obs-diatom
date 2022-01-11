package main

import (
	"database/sql"
	"path"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

/*
 * Construct a database
 */
func NewDB(fpath string) (ObsidianDB, error) {
	db, err := sql.Open("sqlite3", "file:"+fpath+"?_foreign_keys=true&_busy_timeout=5000&_journal_mode=WAL")

	if err != nil {
		return ObsidianDB{}, err
	}

	return ObsidianDB{db, &sync.Mutex{}}, nil
}

// Close a database connection
func (conn *ObsidianDB) Close() error {
	return conn.Db.Close()
}

// Create diatom tables in sqlite
func (conn *ObsidianDB) CreateTables() error {
	// create a file table
	tx, err := conn.Db.Begin()
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

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS metadata (
		file_id  TEXT NOT NULL,
		schema   TEXT NOT NULL,
		content  TEXT NOT NULL,

		PRIMARY KEY(file_id, schema)
	)`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
 * Get the file-hash from the file table in Sqlite
 */
func (conn *ObsidianDB) GetFileHash(fpath string) (string, error) {
	var hash string
	row := conn.Db.QueryRow(`SELECT hash FROM file WHERE file.id = ?`, fpath)

	if row == nil {
		return hash, nil
	}

	err := row.Scan(&hash)
	if err != nil {
		return "", nil
	}

	return hash, nil
}

func (conn *ObsidianDB) InsertFile(fpath, title, hash string) error {
	tx, err := conn.Db.Begin()
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
	tx, err := conn.Db.Begin()
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

	tx, err := conn.Db.Begin()
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

func (conn *ObsidianDB) InsertFrontmatter(frontmatter map[string]interface{}, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bytes, err := yaml.Marshal(frontmatter)

	_, err = tx.Exec(`
	INSERT OR IGNORE INTO metadata (file_id, schema, content) VALUES (?, ?, ?)
	`, fpath, "!frontmatter", string(bytes))

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (conn *ObsidianDB) InsertWikilinks(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.Db.Begin()
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
	return conn.Db.Query(`
		SELECT count(reference) as in_degree, id
				FROM file
			LEFT JOIN wikilink ON file.basename = wikilink.reference
				WHERE id IS NOT NULL
			GROUP BY id
		ORDER BY in_degree DESC`)
}

func (conn *ObsidianDB) GetOutDegree() (*sql.Rows, error) {
	return conn.Db.Query(`
		SELECT count(*) as out_degree, file_id as id
			FROM wikilink
		GROUP BY file_id
		ORDER BY out_degree DESC
	`)
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

func (conn *ObsidianDB) InsertOutDegree(tx *sql.Tx, rows *sql.Rows) error {
	file := struct {
		out_degree int
		id         string
	}{}

	err := rows.Scan(&file.out_degree, &file.id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	UPDATE file
	SET out_degree = ?
	WHERE id = ?
	`, file.out_degree, file.id)

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

	tx, _ := conn.Db.Begin()

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

// Add the file degree
func (conn *ObsidianDB) AddOutDegree() error {
	rows, err := conn.GetOutDegree()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	tx, _ := conn.Db.Begin()

	for rows.Next() {
		if err = conn.InsertOutDegree(tx, rows); err != nil {
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

/*
 * Insert metadata into a database
 */
func (conn *ObsidianDB) InsertMetadata(tx *sql.Tx, fpath, info, yaml string) error {
	_, err := tx.Exec(`
	INSERT OR IGNORE INTO metadata (file_id, schema, content) VALUES (?, ?, ?)
	`, fpath, info, yaml)

	if err != nil {
		return err
	}

	return nil
}

func (conn *ObsidianDB) GetFileIds() ([]string, error) {
	rows, err := conn.Db.Query(`SELECT id from file`)
	fileIds := []string{}
	if err != nil {
		return fileIds, err
	}

	defer rows.Close()

	if err != nil {
		return fileIds, err
	}

	for rows.Next() {
		var fileId string

		err := rows.Scan(&fileId)
		if err != nil {
			return fileIds, err
		}

		fileIds = append(fileIds, fileId)
	}

	return fileIds, nil
}

func (conn *ObsidianDB) DeleteWikilink(fpath string) error {
	_, err := conn.Db.Exec(`DELETE FROM wikilink where file_id = ?`, fpath)
	return err
}

func (conn *ObsidianDB) DeleteTag(fpath string) error {
	_, err := conn.Db.Exec(`DELETE FROM tag where file_id = ?`, fpath)
	return err
}

func (conn *ObsidianDB) DeleteMetadata(fpath string) error {
	_, err := conn.Db.Exec(`DELETE FROM metadata where file_id = ?`, fpath)
	return err
}

func (conn *ObsidianDB) DeleteFile(fpath string) error {
	_, err := conn.Db.Exec(`DELETE FROM file where id = ?`, fpath)
	return err
}
