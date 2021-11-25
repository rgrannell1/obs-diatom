package main

func (conn ObsidianDB) Close() error {
	return conn.db.Close()
}

func (conn ObsidianDB) DropTables() error {
	// create a file table
	_, err := conn.db.Exec(`DROP TABLE IF EXISTS file`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = conn.db.Exec(`DROP TABLE IF EXISTS tag`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = conn.db.Exec(`DROP TABLE IF EXISTS url`)

	if err != nil {
		return err
	}

	_, err = conn.db.Exec(`DROP TABLE IF EXISTS wikilink`)

	if err != nil {
		return err
	}

	return nil
}

func (conn ObsidianDB) CreateTables() error {
	// create a file table
	_, err := conn.db.Exec(`CREATE TABLE IF NOT EXISTS file (
		id    TEXT NOT NULL PRIMARY KEY,
		title TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a tag table
	_, err = conn.db.Exec(`CREATE TABLE IF NOT EXISTS tag (
		tag TEXT      NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	// create a url table
	_, err = conn.db.Exec(`CREATE TABLE IF NOT EXISTS url (
		url      TEXT NOT NULL,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	_, err = conn.db.Exec(`CREATE TABLE IF NOT EXISTS wikilink (
		reference TEXT NOT NULL,
		alias    TEXT,
		file_id  TEXT NOT NULL
	)`)

	if err != nil {
		return err
	}

	return nil
}

func (conn ObsidianDB) InsertFile(fpath string) error {
	// save file to table
	_, err := conn.db.Exec("INSERT INTO file (id, title) VALUES (?, ?)", fpath, fpath)
	return err
}

func (conn ObsidianDB) InsertTags(bodyData *MarkdownData, fpath string) error {
	// save tag to table
	for _, tag := range bodyData.Tags {
		_, err := conn.db.Exec("INSERT INTO tag (tag, file_id) VALUES (?, ?)", tag, fpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn ObsidianDB) InsertUrl(bodyData *MarkdownData, fpath string) error {
	// save url to table
	for _, url := range bodyData.Urls {
		_, err := conn.db.Exec("INSERT INTO url (url, file_id) VALUES (?, ?)", url, fpath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn ObsidianDB) InsertWikilinks(bodyData *MarkdownData, fpath string) error {
	// save wikilinks to table
	for _, wikilink := range bodyData.Wikilinks {
		_, err := conn.db.Exec("INSERT INTO wikilink (reference, alias, file_id) VALUES (?, ?, ?)", wikilink.Reference, wikilink.Alias, fpath)

		if err != nil {
			return err
		}
	}

	return nil
}
