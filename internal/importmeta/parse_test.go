package importmeta

import (
	"testing"
	"time"
)

func TestParseDueAtISO(t *testing.T) {
	got, ok := ParseDueAtISO("2026-06-26T23:59:59Z")
	if !ok {
		t.Fatal("expected ok")
	}
	if got.UTC().Format(time.RFC3339) != "2026-06-26T23:59:59Z" {
		t.Fatalf("got %v", got)
	}
}

func TestParseDueAt(t *testing.T) {
	ref := time.Date(2026, 6, 26, 15, 0, 0, 0, time.Local)

	tests := []struct {
		text string
		want time.Time
		ok   bool
	}{
		{text: "", ok: false},
		{text: "Due No Due Date", ok: false},
		{text: "Due Jun 26 at 11:59pm", want: time.Date(2026, 6, 26, 23, 59, 0, 0, time.Local), ok: true},
		{text: "Due June 26, 2026 at 11:59 PM", want: time.Date(2026, 6, 26, 23, 59, 0, 0, time.Local), ok: true},
		{text: "Due Tomorrow at 11:59pm", want: time.Date(2026, 6, 27, 23, 59, 0, 0, time.Local), ok: true},
		{text: "Due Today at 3pm", want: time.Date(2026, 6, 26, 15, 0, 0, 0, time.Local), ok: true},
		{text: "Due Jan 15", want: time.Date(2026, 1, 15, 23, 59, 59, 0, time.Local), ok: true},
		{text: "Due: Fri Jul 10, 2026 11:59pm", want: time.Date(2026, 7, 10, 23, 59, 0, 0, time.Local), ok: true},
	}

	for _, tc := range tests {
		got, ok := ParseDueAt(tc.text, ref)
		if ok != tc.ok {
			t.Fatalf("ParseDueAt(%q) ok=%v want %v", tc.text, ok, tc.ok)
		}
		if !tc.ok {
			continue
		}
		if !got.Equal(tc.want) {
			t.Fatalf("ParseDueAt(%q) = %v want %v", tc.text, got, tc.want)
		}
	}
}

func TestParsePointsPossible(t *testing.T) {
	tests := []struct {
		text string
		want float64
		ok   bool
	}{
		{text: "", ok: false},
		{text: "25 pts", want: 25, ok: true},
		{text: "30 Points Possible", want: 30, ok: true},
		{text: "100 points possible", want: 100, ok: true},
		{text: "10.5", want: 10.5, ok: true},
		{text: "—", ok: false},
	}

	for _, tc := range tests {
		got, ok := ParsePointsPossible(tc.text)
		if ok != tc.ok {
			t.Fatalf("ParsePointsPossible(%q) ok=%v want %v", tc.text, ok, tc.ok)
		}
		if !tc.ok {
			continue
		}
		if got != tc.want {
			t.Fatalf("ParsePointsPossible(%q) = %v want %v", tc.text, got, tc.want)
		}
	}
}
