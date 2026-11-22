package resource

import "testing"

func TestOrderedChangedFieldsIsStable(t *testing.T) {
	got := orderedChangedFields([]string{"title", "status", "title"})
	want := []string{"status", "title", "title"}
	if len(got) != len(want) {
		t.Fatalf("got %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v", got)
		}
	}
	if ordered := orderedChangedFields(nil); len(ordered) != 0 {
		t.Fatalf("nil fields should become an empty list, got %#v", ordered)
	}
}
