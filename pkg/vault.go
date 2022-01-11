package diatom

import (
	"path/filepath"
	"strings"
)

/*
 * Enumerate all notes in an Obsidian vault
 */
func (vault *ObsidianVault) GetNotes(globSuffix string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(vault.dpath, globSuffix))
	if err != nil {
		return []string{}, err
	}

	filtered := []string{}
	for _, match := range matches {
		if !strings.HasPrefix(match, filepath.Join(vault.dpath, ".trash")) {
			filtered = append(filtered, match)
		}
	}

	return filtered, nil
}
