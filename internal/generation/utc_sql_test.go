package generation

import (
	"os"
	"strings"
	"testing"
)

func TestJobSQLUsesUTCTimestamp(t *testing.T) {
	for _, path := range []string{"../../db/queries/generation.sql", "../../db/queries/media.sql"} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if strings.Contains(text, "CURRENT_TIMESTAMP") {
			t.Fatalf("%s still uses CURRENT_TIMESTAMP", path)
		}
		if !strings.Contains(text, "UTC_TIMESTAMP(6)") {
			t.Fatalf("%s does not use UTC_TIMESTAMP(6)", path)
		}
	}
}
