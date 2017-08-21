package util

import "strings"

var (
	// GoGetSources enumerates all the sources supported
	GoGetSources = map[string]int{
		"github.com":      3,
		"golang.org":      3,
		"code.google.com": 3,
		"code.gitea.io":   2,
		"gopkg.in":        2,
		"bitbucket.org":   3,
	}
)

// NormalizeName splits packages and subpackages
func NormalizeName(name string) (string, string) {
	if len(name) <= 0 {
		return "", ""
	}

	parts := strings.Split(name, "/")
	numParts, ok := GoGetSources[parts[0]]
	if ok {
		if len(parts) > numParts {
			return strings.Join(parts[:3], "/"), strings.Join(parts[3:], "/")
		}
	}

	return name, ""
}
