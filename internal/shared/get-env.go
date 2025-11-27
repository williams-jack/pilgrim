package shared

import (
	"os"
	"regexp"
)

var (
	envRegexPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
)


// ReplaceEnvVariables replaces placeholders in the format ${VAR_NAME} within the input string
// with their corresponding environment variable values. If an environment variable is not set,
// the placeholder is replaced with an empty string.
func ReplaceEnvVariables(input string) string {
	return envRegexPattern.ReplaceAllStringFunc(input, func(match string) string {
		submatches := envRegexPattern.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return ""
		}
		envVarName := submatches[1]
		envVarValue, exists := os.LookupEnv(envVarName)
		if !exists {
			return ""
		}
		return envVarValue
	})
}
