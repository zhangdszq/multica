// Package migrations binds the required schema versions to the compiled source.
package migrations

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed *.up.sql
var files embed.FS

// Versions returns the migration manifest compiled into this revision.
func Versions() ([]string, error) {
	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return nil, err
	}
	versions := make([]string, len(names))
	for i, name := range names {
		versions[i] = strings.TrimSuffix(name, ".up.sql")
	}
	return versions, nil
}
