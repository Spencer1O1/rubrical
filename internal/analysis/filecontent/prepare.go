package filecontent

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"rubrical/internal/config"
)

const defaultMaxFileBytes = config.DefaultDraftMaxFileBytes

type PreparedFile struct {
	FileName string
	MimeType string
	Data     []byte
	Kind     string
	Text     string
}

func Prepare(fileName, mimeType string, data []byte, maxBytes int) (PreparedFile, error) {
	if len(data) == 0 {
		return PreparedFile{}, fmt.Errorf("empty file")
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxFileBytes
	}
	if len(data) > maxBytes {
		return PreparedFile{}, fmt.Errorf("file exceeds %dMB analysis limit", maxBytes>>20)
	}

	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "submission"
	}
	mime := normalizeMime(mimeType, name)

	out := PreparedFile{
		FileName: name,
		MimeType: mime,
		Data:     data,
	}

	switch {
	case isTextMime(mime) || hasTextExtension(name):
		out.Kind = "text"
		out.Text = string(data)
		return out, nil
	case mime == "application/pdf" || strings.EqualFold(filepath.Ext(name), ".pdf"):
		out.Kind = "pdf"
		out.MimeType = "application/pdf"
		return out, nil
	case strings.HasPrefix(mime, "image/"):
		out.Kind = "image"
		return out, nil
	case isDocx(mime, name):
		text, err := extractDocxText(data)
		if err != nil {
			return PreparedFile{}, fmt.Errorf("read docx: %w", err)
		}
		out.Kind = "text"
		out.MimeType = "text/plain"
		out.Text = text
		return out, nil
	default:
		return PreparedFile{}, fmt.Errorf("unsupported file type for analysis: %s", mime)
	}
}

func normalizeMime(mimeType, fileName string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	if mime != "" && mime != "application/octet-stream" {
		return mime
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".txt":
		return "text/plain"
	case ".md", ".markdown":
		return "text/markdown"
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return mimeType
	}
}

func isTextMime(mime string) bool {
	return strings.HasPrefix(mime, "text/") || mime == "application/markdown"
}

func hasTextExtension(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt", ".md", ".markdown":
		return true
	default:
		return false
	}
}

func isDocx(mime, name string) bool {
	if mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		return true
	}
	return strings.EqualFold(filepath.Ext(name), ".docx")
}

func extractDocxText(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		xmlData, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return "", err
		}
		return stripXMLTags(string(xmlData)), nil
	}

	return "", fmt.Errorf("word/document.xml not found")
}

func stripXMLTags(xml string) string {
	var b strings.Builder
	b.Grow(len(xml))
	inTag := false
	for _, r := range xml {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
