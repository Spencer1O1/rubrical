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

	key, err := store.Save(1, 42, "essay.pdf", []byte("pdf bytes"))
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
