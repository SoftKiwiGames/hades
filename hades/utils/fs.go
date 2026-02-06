package utils

import "strings"

func FileHasValidExtension(path string) bool {
	return strings.HasSuffix(path, ".hades.yml") || strings.HasSuffix(path, ".hades.yaml")
}
