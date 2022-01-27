package diatom

import (
	"database/sql"
	"errors"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

/*
 * Construct a database
 */
func NewDB(fpath string) (ObsidianDB, error) {
	db, err := sql.Open("sqlite3", "file:"+fpath+"?_foreign_keys=true&_busy_timeout=5000&_journal_mode=WAL&_sqlite_json=yes")
	return ObsidianDB{db}, err
}

/*
 * Close database connection
 */
func (conn *ObsidianDB) Close() error {
	return conn.Db.Close()
}

/*
 * Create table
 */
func (conn *ObsidianDB) CreateTables() error {
	// create a file table
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`create table if not exists file (
		id         text not null,
		basename   text not null,
		title      text not null,
		hash       text not null,
		in_degree  integer default 0    check(in_degree  >= 0),
		out_degree integer default 0    check(out_degree >= 0),

		primary key(id)
	)`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = tx.Exec(`create table if not exists tag (
		tag      text not null,
		file_id  text not null,

		primary key(tag, file_id)
	)`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = tx.Exec(`create table if not exists url (
		url      text not null,
		file_id  text not null,

		primary key(url, file_id)
	)`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`create table if not exists wikilink (
		reference text not null,
		alias    text,
		file_id  text not null,

		primary key(reference, alias, file_id)
	)`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`create table if not exists metadata (
		file_id  text not null,
		schema   text not null,
		content  text not null,

		primary key(file_id, schema)
	)`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`create table if not exists heading (
		heading  text not null,
		level    integer not null,
		file_id  text not null,

		primary key(heading, level, file_id)
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

/*
 *
 */
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
	insert or ignore into file (id, basename, title, hash) values (?, ?, ?, ?)
	on conflict (id)
	do update set title = ?, basename = ?, hash = ?
	`, fpath, basename, title, hash, title, basename, hash)

	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) InsertTags(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, tag := range bodyData.Tags {
		_, err := tx.Exec(`insert or replace into tag (tag, file_id) values (?, ?)`, tag, fpath)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) InsertUrl(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range bodyData.Urls {
		_, err := tx.Exec(`
		insert or ignore into url (url, file_id) values (?, ?)
		`, url, fpath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) InsertFrontmatter(frontmatter map[string]interface{}, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return err
	}

	json, err := yamlToJson(string(bytes))
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	insert or replace into metadata (file_id, schema, content) values (?, ?, ?)
	`, fpath, "!frontmatter", json)

	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) InsertWikilinks(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, wikilink := range bodyData.Wikilinks {
		_, err := tx.Exec(`
		insert or ignore into wikilink (reference, alias, file_id) values (?, ?, ?)
		`, wikilink.Reference, wikilink.Alias, fpath)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) GetInDegree() (*sql.Rows, error) {
	return conn.Db.Query(`
		select count(reference) as in_degree, id
				from file
			left join wikilink on file.basename = wikilink.reference
				where id is not null
			group by id
		order by in_degree desc`)
}

func (conn *ObsidianDB) GetOutDegree() (*sql.Rows, error) {
	return conn.Db.Query(`
		select count(*) as out_degree, file_id as id
			from wikilink
		group by file_id
		order by out_degree desc
	`)
}

/*
 * Insert a degree row into the database
 */
func (conn *ObsidianDB) InsertInDegree(rows *sql.Rows) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}

	file := struct {
		in_degree int
		id        string
	}{}

	err = rows.Scan(&file.in_degree, &file.id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	update file
	set in_degree = ?
	where id = ?
	`, file.in_degree, file.id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) InsertOutDegree(rows *sql.Rows) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}

	file := struct {
		out_degree int
		id         string
	}{}

	err = rows.Scan(&file.out_degree, &file.id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	update file
	set out_degree = ?
	where id = ?
	`, file.out_degree, file.id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) AddInDegree() error {
	rows, err := conn.GetInDegree()
	if err != nil {
		return err
	}

	count := 0

	for rows.Next() {
		count++
		if err = conn.InsertInDegree(rows); err != nil {
			return err
		}
	}

	if count == 0 {
		return errors.New("no files present in database")
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	return nil
}

/*
 *
 */
func (conn *ObsidianDB) AddOutDegree() error {
	rows, err := conn.GetOutDegree()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	count := 0

	for rows.Next() {
		count++
		if err = conn.InsertOutDegree(rows); err != nil {
			return err
		}
	}

	if count == 0 {
		return errors.New("no files present in database")
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
	insert or ignore into metadata (file_id, schema, content) values (?, ?, ?)
	`, fpath, info, yaml)

	return err
}

/*
 * Insert headings into sqlite
 */
func (conn *ObsidianDB) InsertHeadings(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, heading := range bodyData.Headings {
		_, err := tx.Exec(`insert or replace into heading (heading, level, file_id) values (?, ?, ?)`, heading.Text, heading.Level, fpath)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

/*
 *
 */
func (conn *ObsidianDB) GetFileIds() ([]string, error) {
	rows, err := conn.Db.Query(`select id from file`)
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

/*
 *
 */
func (conn *ObsidianDB) DeleteWikilink(fpath string) error {
	_, err := conn.Db.Exec(`delete from wikilink where file_id = ?`, fpath)
	return err
}

/*
 *
 */
func (conn *ObsidianDB) DeleteTag(fpath string) error {
	_, err := conn.Db.Exec(`delete from tag where file_id = ?`, fpath)
	return err
}

/*
 *
 */
func (conn *ObsidianDB) DeleteMetadata(fpath string) error {
	_, err := conn.Db.Exec(`delete from metadata where file_id = ?`, fpath)
	return err
}

/*
 *
 */
func (conn *ObsidianDB) DeleteFile(fpath string) error {
	_, err := conn.Db.Exec(`delete from file where id = ?`, fpath)
	return err
}
