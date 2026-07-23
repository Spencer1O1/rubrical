package files_test

import (
	"strings"
	"testing"

	"rubrical/internal/analysispipeline/files"
)

func TestCanInspectKind_matchesRoute(t *testing.T) {
	cases := []struct {
		provider string
		kind     files.FileKind
		want     bool
	}{
		{"openai", files.KindImage, true},
		{"openai", files.KindOffice, true},
		{"openai", files.KindSpreadsheet, true},
		{"openai", files.KindMedia, false},
		{"anthropic", files.KindImage, true},
		{"anthropic", files.KindPDF, true},
		{"anthropic", files.KindDocx, true},
		{"anthropic", files.KindText, true},
		{"anthropic", files.KindOffice, false},
		{"anthropic", files.KindSpreadsheet, false},
		{"anthropic", files.KindMedia, false},
	}
	for _, tc := range cases {
		if got := files.CanInspectKind(tc.provider, tc.kind); got != tc.want {
			t.Fatalf("%s kind=%v: got %v want %v", tc.provider, tc.kind, got, tc.want)
		}
	}
}

func TestPromptCapabilities_derivedFromKinds(t *testing.T) {
	oCan, oCannot := files.PromptCapabilities("openai")
	aCan, aCannot := files.PromptCapabilities("anthropic")

	if !containsExact(oCan, "Office (pptx, odp, odt, rtf, …)") {
		t.Fatalf("openai should inspect Office: can=%v", oCan)
	}
	if !containsExact(oCan, "Excel (xlsx/xls)") {
		t.Fatalf("openai should inspect Excel: can=%v", oCan)
	}
	if containsExact(aCan, "Office (pptx, odp, odt, rtf, …)") {
		t.Fatalf("anthropic should not inspect Office: can=%v", aCan)
	}
	if containsExact(aCan, "Excel (xlsx/xls)") {
		t.Fatalf("anthropic should not claim Excel: can=%v", aCan)
	}
	if !containsExact(aCan, "text files (txt, csv, tsv, …)") {
		t.Fatalf("anthropic can inspect text/csv: can=%v", aCan)
	}
	if !containsExact(aCannot, "Office (pptx, odp, odt, rtf, …)") {
		t.Fatalf("anthropic cannot should list Office: %v", aCannot)
	}
	for _, list := range [][]string{oCannot, aCannot} {
		if !containsExact(list, "video and audio") {
			t.Fatalf("nobody inspects video yet: %v", list)
		}
	}
}

func TestClassify_csvIsTextNotSpreadsheet(t *testing.T) {
	if got := files.Classify("grades.csv", "text/csv", []byte("a,b\n1,2\n")); got != files.KindText {
		t.Fatalf("csv kind = %v, want KindText", got)
	}
	if got := files.Classify("book.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []byte("PK")); got != files.KindSpreadsheet {
		t.Fatalf("xlsx kind = %v, want KindSpreadsheet", got)
	}
}

func TestFormatPromptCapabilities(t *testing.T) {
	s := files.FormatPromptCapabilities("anthropic")
	if !strings.Contains(s, "Can:") || !strings.Contains(s, "Cannot:") {
		t.Fatalf("bad format: %s", s)
	}
	if strings.Contains(s, "live/in-person") {
		t.Fatal("locus belongs in the rule, not capabilities")
	}
}

func containsExact(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
