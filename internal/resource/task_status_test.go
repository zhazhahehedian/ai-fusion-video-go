package resource

import (
	"database/sql"
	"testing"
)

func TestDeletableTaskStatusMatchesJava(t *testing.T) {
	if DeletableTaskStatus(sql.NullInt32{}) {
		t.Fatal("null status is not finished")
	}
	for _, status := range []int32{0, 1, 5, 7} {
		if DeletableTaskStatus(sql.NullInt32{Int32: status, Valid: true}) {
			t.Fatalf("status %d must stay", status)
		}
	}
	for _, status := range []int32{2, 3, 4, 6} {
		if !DeletableTaskStatus(sql.NullInt32{Int32: status, Valid: true}) {
			t.Fatalf("status %d can be deleted", status)
		}
	}
}
