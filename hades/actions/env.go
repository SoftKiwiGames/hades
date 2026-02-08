package actions

import (
	"regexp"
)

var envVarPatternShared = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpandEnvVars expands ${VAR} references in a string using the provided environment
// If a variable is not found in env, it's left as-is (${VAR})
func ExpandEnvVars(s string, env map[string]string) string {
	return envVarPatternShared.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR}
		varName := match[2 : len(match)-1]

		// Look up in the provided environment
		if value, ok := env[varName]; ok {
			return value
		}

		// If not found, return the original ${VAR}
		return match
	})
}
