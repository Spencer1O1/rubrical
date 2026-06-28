package filecontent

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestPrepare_textFile(t *testing.T) {
	prepared, err := Prepare("notes.txt", "text/plain", []byte("hello draft"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Kind != "text" || prepared.Text != "hello draft" {
		t.Fatalf("unexpected prepared file: %+v", prepared)
	}
}

func TestPrepare_docx(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("<w:document><w:body><w:p><w:r><w:t>Hello docx</w:t></w:r></w:p></w:body></w:document>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	prepared, err := Prepare("essay.docx", "", buf.Bytes(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Kind != "text" {
		t.Fatalf("kind = %q", prepared.Kind)
	}
	if prepared.Text == "" {
		t.Fatal("expected extracted docx text")
	}
}

func TestPrepare_rejectsEmpty(t *testing.T) {
	if _, err := Prepare("empty.txt", "text/plain", nil, 0); err == nil {
		t.Fatal("expected error")
	}
}
