package parse

import (
	"strings"
	"testing"
)

func TestParseSQLFile_DescriptionDirectiveSingleLine(t *testing.T) {
	sql := `-- $pilgrim:[description] Add users table
CREATE TABLE users (id INT);`
	r := strings.NewReader(sql)
	result, err := ParseSQLFile(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Directives.Description != "Add users table" {
		t.Errorf("unexpected Description: %q", result.Directives.Description)
	}
}

func TestParseSQLFile_DescriptionDirectiveMultiLine(t *testing.T) {
	sql := `-- $pilgrim:[description] Add users table
-- $pilgrim:[description] with extra info
CREATE TABLE users (id INT);`
	r := strings.NewReader(sql)
	result, err := ParseSQLFile(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Add users table with extra info"
	if result.Directives.Description != expected {
		t.Errorf("unexpected Description: %q", result.Directives.Description)
	}
}

func TestParseSQLFile_DirectivesAfterSQLIgnored(t *testing.T) {
	sql := `SELECT 1;
-- $pilgrim:[description] ignored`
	r := strings.NewReader(sql)
	result, err := ParseSQLFile(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Directives.Description != "" {
		t.Errorf("expected Description to be empty")
	}
}
