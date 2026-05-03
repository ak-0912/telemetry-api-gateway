package ginhttp

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseRFC3339Like_NanoAndPlain(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"2024-01-02T15:04:05Z", "2024-01-02T15:04:05Z"},
		{"2024-01-02T15:04:05.123456789Z", "2024-01-02T15:04:05.123456789Z"},
	}
	for _, tc := range cases {
		got, err := parseRFC3339Like(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		want, _ := time.Parse(time.RFC3339Nano, tc.want)
		if !got.Equal(want) {
			t.Fatalf("%q: got %v want %v", tc.in, got, want)
		}
	}
}

func TestParseRFC3339Like_Invalid(t *testing.T) {
	t.Parallel()
	_, err := parseRFC3339Like("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRFC3339Like_SpaceAndPlus00(t *testing.T) {
	t.Parallel()
	cases := []string{
		"2026-05-01 08:00:00+00",
		"2026-05-01T08:00:00+00",
		"2026-05-01T08:00:00+0000",
	}
	for _, in := range cases {
		got, err := parseRFC3339Like(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		want, _ := time.Parse(time.RFC3339, "2026-05-01T08:00:00Z")
		if !got.Equal(want) {
			t.Fatalf("%q: got %v want %v", in, got, want)
		}
	}
}

func TestParseTimeWindow_Query(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/x?start_time=2026-05-01T14:15:00Z&end_time=2026-05-03T06:30:00Z", nil)
	start, end, err := parseTimeWindow(req)
	if err != nil {
		t.Fatal(err)
	}
	if start == nil || end == nil {
		t.Fatal("expected both")
	}
	if start.Format(time.RFC3339) != "2026-05-01T14:15:00Z" {
		t.Fatalf("start %v", start)
	}
	if end.Format(time.RFC3339) != "2026-05-03T06:30:00Z" {
		t.Fatalf("end %v", end)
	}
}
