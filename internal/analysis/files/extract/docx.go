package extract

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func Docx(data []byte) (string, error) {
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

func Text(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}
	if utf8.Valid(data) {
		return string(data), nil
	}
	return "", fmt.Errorf("could not decode as UTF-8 text")
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
