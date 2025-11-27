package parse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	directiveRegexpStr = "^-- \\$pilgrim\\:\\[(description)](.*)$"
)

// SQLWithDirectives holds the SQL content along with any parsed directives.
type SQLWithDirectives struct {
	Sql        string
	Directives *PilgrimDirectives
}

// Directives in pilgrim add either metadata or instruct the migrator to change
// ts behavior. The following directives are supported:
//
// Description Directive:
// -- $pilgrim:[description] This is a description of the migration. It can be
// -- $pilgrim:[description] multi-line if needed.
//
// Provides a human-readable description of the migration. This description is
// stored in the migration history table and is used for documentation purposes.
type PilgrimDirectives struct {
	Description string
}

// ParseSQLFile reads SQL content from the provided reader, extracts any
// directives, and returns a SQLWithDirectives struct pointer along with any
// parsing errors. Only directives at the beginning of the file are considered;
// directives appearing after SQL statements are ignored.
//
// Note that "parsing" in this context only refers to the extraction of directives.
// The actual SQL syntax is not validated.
func ParseSQLFile(reader io.Reader) (*SQLWithDirectives, error) {
	var swd SQLWithDirectives
	var parseErrors []error
	directiveRegexp := regexp.MustCompile(directiveRegexpStr)
	descriptionLines := []string{}
	sqlLines := []string{}
	directives := PilgrimDirectives{}
	scanner := bufio.NewScanner(reader)
	inDirectives := true
	lineNumber := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNumber++
		if line == "" {
			continue
		}
		if inDirectives {
			matches := directiveRegexp.FindStringSubmatch(line)
			if len(matches) > 0 {
				directiveType := matches[1]
				directiveValue := strings.TrimSpace(matches[2])
				switch directiveType {
				case "description":
					descriptionLines = append(descriptionLines, directiveValue)
				}
				continue
			}
			inDirectives = false
		}
		sqlLines = append(sqlLines, scanner.Text())
	}
	if scanner.Err() != nil {
		parseErrors = append(parseErrors, scanner.Err())
	}
	if len(parseErrors) > 0 {
		errStrs := ""
		for _, e := range parseErrors {
			errStrs += e.Error() + "\n"
		}
		return nil, fmt.Errorf("errors parsing SQL file:\n%s", errStrs)
	}
	directives.Description = strings.Join(descriptionLines, " ")
	swd.Directives = &directives
	swd.Sql = strings.Join(sqlLines, "\n")
	return &swd, nil
}
