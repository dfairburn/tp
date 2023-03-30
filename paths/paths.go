package paths

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func OpenOrCreateFile(p string) *os.File {
	// check if we can open the file for writing straight away
	expandedPath := Expand(p)
	lf, err := os.Create(expandedPath)
	if err != nil {
		// create dir so we can open the file for writing
		dir := path.Dir(p)
		expandedDir := Expand(dir)

		err := os.Mkdir(expandedDir, 0750)
		if err != nil {
			panic(fmt.Errorf("could not create log file path: %v", err))
		}

		lf, err = os.Create(expandedPath)
		if err != nil {
			panic(fmt.Errorf("could not open log file path: %v", err))
		}
	}

	return lf
}

// Expand resolves relative paths to absolute paths
func Expand(p string) string {
	tilde := expandTilde(p)
	local := expandLocal(p)

	if local != p {
		return local
	}

	if tilde != p {
		return tilde
	}

	return p
}

func expandTilde(p string) string {
	return expand("~/", "$HOME", p)
}

func expandLocal(p string) string {
	return expand("./", "$PWD", p)
}

func expand(prefix, elem, s string) string {
	if strings.HasPrefix(s, prefix) {
		replaced := filepath.Join(elem, s[2:])
		return os.ExpandEnv(replaced)
	}
	return s

}
