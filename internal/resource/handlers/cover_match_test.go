package handlers

import (
	"database/sql"
	"testing"
)

func TestCoverMatchesOnlySameNonEmptyURL(t *testing.T) {
	image := sql.NullString{String: "https://cdn.example/a.png", Valid: true}
	cover := image
	if !coverMatches(image, cover) {
		t.Fatal("same url should clear the cover")
	}
	if coverMatches(image, sql.NullString{String: "https://cdn.example/b.png", Valid: true}) {
		t.Fatal("different url must keep the cover")
	}
	if coverMatches(sql.NullString{}, cover) || coverMatches(image, sql.NullString{}) {
		t.Fatal("missing url must not clear the cover")
	}
	blank := sql.NullString{String: "", Valid: true}
	if coverMatches(blank, blank) {
		t.Fatal("empty url is not a cover match")
	}
}
