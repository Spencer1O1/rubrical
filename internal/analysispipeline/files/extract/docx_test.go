package extract

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestDocx(t *testing.T) {
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

	text, err := Docx(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Fatal("expected extracted docx text")
	}
}

func TestText_rejectsEmpty(t *testing.T) {
	if _, err := Text(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestText_utf8(t *testing.T) {
	text, err := Text([]byte("hello draft"))
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello draft" {
		t.Fatalf("got %q", text)
	}
}
