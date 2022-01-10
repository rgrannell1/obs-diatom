package diatom

import (
	"path/filepath"
)

/*
 * Enumerate all notes in an Obsidian vault
 */
func (vault *ObsidianVault) GetNotes(globSuffix string) ([]string, error) {
	return filepath.Glob(vault.dpath + "/" + globSuffix)
}
