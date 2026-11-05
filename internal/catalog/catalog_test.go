package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalog(t *testing.T) {
	result, err := Inspect(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalRules != 21 || result.TotalPortHints != 15 || len(result.Protocols) != 11 {
		t.Fatalf("%+v", result)
	}
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i-1].Name >= result.Entries[i].Name {
			t.Fatal("entries not sorted")
		}
	}
}

func TestInvalidCatalogs(t *testing.T) {
	for _, tc := range []struct{ content, message string }{
		{`{"rules":[{"name":"a","protocol":"X","pattern":"x"},{"name":"a","protocol":"Y","pattern":"y"}]}`, "duplicate"},
		{`{"rules":[{"protocol":"X","pattern":"x"}]}`, "empty"},
		{`{"rules":[{"name":"a","protocol":"X","pattern":"["}]}`, "compile"},
	} {
		path := filepath.Join(t.TempDir(), "rules.json")
		if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Inspect(path); err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("wanted %s, got %v", tc.message, err)
		}
	}
}
