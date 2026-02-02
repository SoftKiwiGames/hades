package actions

import (
	"fmt"
	"regexp"
	"strings"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// expandEnv expands ${VAR} references in a string using the provided env map
func expandEnv(s string, env map[string]string) (string, error) {
	var missingVars []string

	result := envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR}
		varName := match[2 : len(match)-1]

		value, ok := env[varName]
		if !ok {
			missingVars = append(missingVars, varName)
			return match
		}
		return value
	})

	if len(missingVars) > 0 {
		return "", fmt.Errorf("missing environment variables: %s", strings.Join(missingVars, ", "))
	}

	return result, nil
}
