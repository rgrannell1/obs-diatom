package main

func (conn ObsidianDB) Close() error {
	return conn.db.Close()
}

func (conn ObsidianDB) DropTables() error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DROP TABLE IF EXISTS file`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = tx.Exec(`DROP TABLE IF EXISTS tag`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = tx.Exec(`DROP TABLE IF EXISTS url`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`DROP TABLE IF EXISTS wikilink`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (conn ObsidianDB) CreateTables() error {
	// create a file table
	tx, err := conn.db.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS file (
		id    TEXT NOT NULL PRIMARY KEY,
		title TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS tag (
		tag TEXT      NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS url (
		url      TEXT NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS wikilink (
		reference TEXT NOT NULL,
		alias    TEXT,
		file_id  TEXT NOT NULL
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

func (conn ObsidianDB) InsertFile(fpath string, title string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// save file to table
	_, err = tx.Exec("INSERT INTO file (id, title) VALUES (?, ?)", fpath, title)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (conn ObsidianDB) InsertTags(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, tag := range bodyData.Tags {
		_, err := tx.Exec("INSERT INTO tag (tag, file_id) VALUES (?, ?)", tag, fpath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (conn ObsidianDB) InsertUrl(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range bodyData.Urls {
		_, err := tx.Exec("INSERT INTO url (url, file_id) VALUES (?, ?)", url, fpath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (conn ObsidianDB) InsertWikilinks(bodyData *MarkdownData, fpath string) error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, wikilink := range bodyData.Wikilinks {
		_, err := tx.Exec("INSERT INTO wikilink (reference, alias, file_id) VALUES (?, ?, ?)", wikilink.Reference, wikilink.Alias, fpath)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
