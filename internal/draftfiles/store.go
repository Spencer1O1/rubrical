package draftfiles

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	root string
}

func NewStore(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &Store{root: abs}, nil
}

func (s *Store) Save(userID, assignmentID int64, filename string, data []byte) (storageKey string, err error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}

	ext := sanitizeExtension(filepath.Ext(filename))
	id, err := randomID()
	if err != nil {
		return "", err
	}

	storageKey = filepath.ToSlash(filepath.Join(
		"drafts",
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("%d", assignmentID),
		id+ext,
	))
	fullPath := filepath.Join(s.root, filepath.FromSlash(storageKey))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create draft dir: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write draft file: %w", err)
	}

	return storageKey, nil
}

func (s *Store) Delete(storageKey string) error {
	if strings.TrimSpace(storageKey) == "" {
		return nil
	}

	fullPath := filepath.Join(s.root, filepath.FromSlash(storageKey))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete draft file: %w", err)
	}

	dir := filepath.Dir(fullPath)
	if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
		_ = os.Remove(dir)
	}

	return nil
}

func (s *Store) Read(storageKey string) ([]byte, error) {
	if strings.TrimSpace(storageKey) == "" {
		return nil, fmt.Errorf("empty storage key")
	}
	fullPath := filepath.Join(s.root, filepath.FromSlash(storageKey))
	return os.ReadFile(fullPath)
}

func (s *Store) Path(storageKey string) string {
	return filepath.Join(s.root, filepath.FromSlash(storageKey))
}

func sanitizeExtension(ext string) string {
	ext = strings.ToLower(ext)
	if ext == "" || len(ext) > 16 {
		return ".bin"
	}
	for _, r := range ext[1:] {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return ".bin"
		}
	}
	return ext
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("random id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
