package files

import "testing"

func TestRouteFileRejectsCorruptPDF(t *testing.T) {
	_, attachment, note := routeFile("openai", RawFile{
		Path:     LogicalPath{RelativePath: "essay.pdf"},
		FileName: "essay.pdf",
		MimeType: "application/pdf",
		Data:     []byte("[object Object]"),
	})

	if attachment != nil {
		t.Fatal("expected no attachment for corrupt pdf")
	}
	if note == "" {
		t.Fatal("expected skip note for corrupt pdf")
	}
}
