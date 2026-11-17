package handlers

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

func TestOptionalNullStringRejectsWrongType(t *testing.T) {
	got, err := optionalNullString(123)
	if err == nil || got.Valid {
		t.Fatalf("wrong type must be a validation error, got (%v, %v)", got, err)
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.HTTPStatus != 400 {
		t.Fatalf("want apperr 400, got %v", err)
	}

	empty, err := optionalNullString("  ")
	if err != nil || empty.Valid {
		t.Fatalf("blank string clears the field, got (%v, %v)", empty, err)
	}
	text, err := optionalNullString(" 标题 ")
	if err != nil || !text.Valid || text.String != "标题" {
		t.Fatalf("trimmed string = (%v, %v)", text, err)
	}
	missing, err := optionalNullString(nil)
	if err != nil || missing.Valid {
		t.Fatalf("nil stays null, got (%v, %v)", missing, err)
	}
}

func TestOptionalNullInt32RejectsFractionAndOverflow(t *testing.T) {
	if got, err := optionalNullInt32(1.7); err == nil || got.Valid {
		t.Fatalf("1.7 must not truncate, got (%v, %v)", got, err)
	}
	if got, err := optionalNullInt32(1e10); err == nil || got.Valid {
		t.Fatalf("1e10 must not wrap int32, got (%v, %v)", got, err)
	}
	if got, err := optionalNullInt32("12"); err == nil || got.Valid {
		t.Fatalf("string must not become null, got (%v, %v)", got, err)
	}
	got, err := optionalNullInt32(float64(12))
	if err != nil || !got.Valid || got.Int32 != 12 {
		t.Fatalf("integral float = (%v, %v)", got, err)
	}
	missing, err := optionalNullInt32(nil)
	if err != nil || missing.Valid {
		t.Fatalf("nil stays null, got (%v, %v)", missing, err)
	}
}

func TestOptionalJSONDoesNotSwallowMarshalError(t *testing.T) {
	if raw, err := optionalJSON(make(chan int)); err == nil || string(raw) == "{}" {
		t.Fatalf("marshal error must not become {}, got (%s, %v)", raw, err)
	}
	raw, err := optionalJSON(nil)
	if err != nil || string(raw) != "{}" {
		t.Fatalf("nil object = (%s, %v)", raw, err)
	}
	raw, err = optionalJSON(map[string]any{"a": 1})
	if err != nil || string(raw) != `{"a":1}` {
		t.Fatalf("object = (%s, %v)", raw, err)
	}

	if raw, err := optionalJSONArray(make(chan int)); err == nil || string(raw) == "[]" {
		t.Fatalf("marshal error must not become [], got (%s, %v)", raw, err)
	}
	raw, err = optionalJSONArray(nil)
	if err != nil || string(raw) != "[]" {
		t.Fatalf("nil array = (%s, %v)", raw, err)
	}
}

func TestOptionalNullInt64RejectsWrongType(t *testing.T) {
	if got, err := optionalNullInt64(1.7); err == nil || got.Valid {
		t.Fatalf("fraction must not clear the field, got (%v, %v)", got, err)
	}
	if got, err := optionalNullInt64(true); err == nil || got.Valid {
		t.Fatalf("bool must not clear the field, got (%v, %v)", got, err)
	}
	got, err := optionalNullInt64(int64(9))
	if err != nil || !got.Valid || got.Int64 != 9 {
		t.Fatalf("id = (%v, %v)", got, err)
	}
}

func TestAsInt64ValueRejectsNonIntegral(t *testing.T) {
	if _, err := asInt64Value(1.7); err == nil {
		t.Fatal("1.7 accepted")
	}
	if _, err := asInt64Value(math.NaN()); err == nil {
		t.Fatal("NaN accepted")
	}
	got, err := asInt64Value(float64(1))
	if err != nil || got != 1 {
		t.Fatalf("1.0 = (%d, %v)", got, err)
	}
	got, err = asInt64Value(json.Number("8"))
	if err != nil || got != 8 {
		t.Fatalf("json number = (%d, %v)", got, err)
	}
	if _, err := asInt64Value(json.Number("1.7")); err == nil {
		t.Fatal("json number 1.7 accepted")
	}
	if _, err := asInt64Value(json.Number("9223372036854775808")); err == nil {
		t.Fatal("json number overflow accepted")
	}
}

func TestRequiredIntRejectsOverflow(t *testing.T) {
	if _, err := requiredPositiveInt(1.7, "episodeNumber"); err == nil {
		t.Fatal("fractional episodeNumber accepted")
	}
	if _, err := requiredPositiveInt(1e10, "episodeNumber"); err == nil {
		t.Fatal("1e10 episodeNumber accepted")
	}
	if _, err := requiredPositiveInt(float64(0), "episodeNumber"); err == nil {
		t.Fatal("zero episodeNumber accepted")
	}
	got, err := requiredPositiveInt(float64(math.MaxInt32), "episodeNumber")
	if err != nil || got != math.MaxInt32 {
		t.Fatalf("max int32 = (%d, %v)", got, err)
	}

	if _, err := requiredNonNegInt(float64(-1), "sortOrder"); err == nil {
		t.Fatal("negative sortOrder accepted")
	}
	if _, err := requiredNonNegInt(float64(math.MaxInt32)+1, "sortOrder"); err == nil {
		t.Fatal("overflow sortOrder accepted")
	}
	got, err = requiredNonNegInt(float64(0), "sortOrder")
	if err != nil || got != 0 {
		t.Fatalf("zero sortOrder = (%d, %v)", got, err)
	}

	if _, err := requiredPositiveID(1.5, "scriptId"); err == nil {
		t.Fatal("fractional scriptId accepted")
	}
	id, err := requiredPositiveID(float64(42), "scriptId")
	if err != nil || id != 42 {
		t.Fatalf("scriptId = (%d, %v)", id, err)
	}
}
