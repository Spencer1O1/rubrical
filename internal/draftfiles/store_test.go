package draftfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSaveDelete(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.Save(1, 42, "essay.pdf", []byte("%PDF-1.4\npdf bytes"))
	if err != nil {
		t.Fatal(err)
	}

	fullPath := store.Path(key)
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("file missing: %v", err)
	}

	if err := store.Delete(key); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Fatal("file should be deleted")
	}
}

func TestStoreSaveSanitizesExtension(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.Save(1, 1, `weird"name`, []byte("x"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(key) != ".bin" {
		t.Fatalf("expected .bin extension, got %q", key)
	}
}

func TestSaveRejectsCorruptPDF(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Save(1, 1, "essay.pdf", []byte("[object Object]"))
	if err == nil {
		t.Fatal("expected error for corrupt pdf")
	}
}

func TestSaveAcceptsValidPDF(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.Save(1, 1, "essay.pdf", []byte("%PDF-1.4\n"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(key)))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "%PDF-1.4\n" {
		t.Fatalf("stored=%q", string(data))
	}
}
