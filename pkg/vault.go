package diatom

import (
	"path/filepath"
	"strings"
)

/*
 * Enumerate all notes in an Obsidian vault.
 *
 */
func (vault *ObsidianVault) GetNotes() ([]string, error) {
	flatMatches, err := filepath.Glob(filepath.Join(vault.dpath, "*.md"))
	if err != nil {
		return []string{}, err
	}

	nestedMatches, err := filepath.Glob(filepath.Join(vault.dpath, "**/*.md"))
	if err != nil {
		return []string{}, err
	}

	filtered := []string{}
	for _, match := range append(nestedMatches, flatMatches...) {
		if !strings.HasPrefix(match, filepath.Join(vault.dpath, ".trash")) {
			filtered = append(filtered, match)
		}
	}

	return filtered, nil
}
