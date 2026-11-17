package auth

import "testing"

func TestDeserializeSessionJSONAndLegacy(t *testing.T) {
	s, ok := deserializeSession(`{"userId":7,"username":"bob","currentTeamId":null}`)
	if !ok || s.UserID != 7 || s.Username != "bob" || s.CurrentTeamID != nil {
		t.Fatalf("json session: %+v ok=%v", s, ok)
	}

	teamID := int64(3)
	rawTeam := `{"userId":7,"username":"bob","currentTeamId":3}`
	s, ok = deserializeSession(rawTeam)
	if !ok || s.CurrentTeamID == nil || *s.CurrentTeamID != teamID {
		t.Fatalf("json with team: %+v ok=%v", s, ok)
	}

	s, ok = deserializeSession("9:alice")
	if !ok || s.UserID != 9 || s.Username != "alice" {
		t.Fatalf("legacy session: %+v ok=%v", s, ok)
	}

	if _, ok := deserializeSession(""); ok {
		t.Fatal("empty should fail")
	}
}
