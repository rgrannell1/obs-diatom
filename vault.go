package diatom

import (
	"path/filepath"
)

func (vault *ObsidianVault) GetNotes(globSuffix string) ([]string, error) {
	return filepath.Glob(vault.dpath + "/" + globSuffix)
}
