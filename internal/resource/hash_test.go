package resource

import "testing"

func TestHashCommandStableAndSensitive(t *testing.T) {
	cmd := CreateCommand{
		ReqID:   "req-1",
		ResType: TypeScriptEpisode,
		Fields:  map[string]any{"scriptId": 1, "episodeNumber": 1, "parentVersion": 0},
	}
	h1, err := HashCommand(cmd, 9, OriginUser)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashCommand(cmd, 9, OriginUser)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("same command should hash equal: %s vs %s", h1, h2)
	}
	changed := cmd
	changed.Fields = map[string]any{"scriptId": 1, "episodeNumber": 2, "parentVersion": 0}
	h3, err := HashCommand(changed, 9, OriginUser)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h3 {
		t.Fatal("different body should change hash")
	}
	h4, err := HashCommand(cmd, 10, OriginUser)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h4 {
		t.Fatal("different actor should change hash")
	}
}

func TestValidateRequestID(t *testing.T) {
	if _, err := ValidateRequestID(""); err == nil {
		t.Fatal("empty should fail")
	}
	if _, err := ValidateRequestID("ok-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateRequestID("bad id"); err == nil {
		t.Fatal("space should fail")
	}
}
