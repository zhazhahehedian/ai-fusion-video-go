package platform

import "testing"

func TestWithUTCTimeZone(t *testing.T) {
	got := WithUTCTimeZone("user:pass@tcp(127.0.0.1:3306)/ai_fusion_video?parseTime=true&loc=UTC")
	want := "user:pass@tcp(127.0.0.1:3306)/ai_fusion_video?parseTime=true&loc=UTC&time_zone=%27%2B00%3A00%27"
	if got != want {
		t.Fatalf("got %s", got)
	}
	kept := "user:pass@tcp(127.0.0.1:3306)/db?time_zone=%27%2B08%3A00%27"
	if WithUTCTimeZone(kept) != kept {
		t.Fatal("explicit session time zone must stay")
	}
	if WithUTCTimeZone("user:pass@tcp(127.0.0.1:3306)/db") != "user:pass@tcp(127.0.0.1:3306)/db?time_zone=%27%2B00%3A00%27" {
		t.Fatal("dsn without query should gain a question mark")
	}
}
