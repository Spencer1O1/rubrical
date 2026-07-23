package files

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
)

type zipExpandState struct {
	limits          Limits
	depth           int
	uncompressedSum int64
}

func expandZip(raw RawFile, limits Limits) ([]RawFile, []string, error) {
	state := zipExpandState{limits: limits.withDefaults()}
	archiveRoot := raw.Path.String()
	if archiveRoot == "" {
		archiveRoot = raw.FileName
	}
	return state.expand(raw.Data, archiveRoot, 0)
}

func (s *zipExpandState) expand(data []byte, archiveRoot string, depth int) ([]RawFile, []string, error) {
	if depth > s.limits.zipMaxDepth() {
		return nil, []string{fmt.Sprintf("%s: zip depth limit exceeded", archiveRoot)}, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: corrupt zip archive", archiveRoot)}, nil
	}

	var leaves []RawFile
	var notes []string

	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		name := cleanZipEntryPath(entry.Name)
		if name == "" {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (invalid path)", archiveRoot, entry.Name))
			continue
		}

		entrySize := int64(entry.UncompressedSize64)
		if entrySize > int64(s.limits.MaxUploadBytes) {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (upload size limit)", archiveRoot, name))
			continue
		}
		if s.uncompressedSum+entrySize > int64(s.limits.zipMaxUncompressedBytes()) {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (total uncompressed size limit)", archiveRoot, name))
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (encrypted or unreadable)", archiveRoot, name))
			continue
		}
		entryData, err := io.ReadAll(io.LimitReader(rc, int64(s.limits.MaxUploadBytes)+1))
		rc.Close()
		if err != nil {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (read error)", archiveRoot, name))
			continue
		}
		if len(entryData) == 0 {
			continue
		}
		if len(entryData) > s.limits.MaxUploadBytes {
			notes = append(notes, fmt.Sprintf("%s/%s: skipped (upload size limit)", archiveRoot, name))
			continue
		}

		s.uncompressedSum += int64(len(entryData))

		baseName := path.Base(name)
		logical := LogicalPath{ArchiveRoot: archiveRoot, RelativePath: name}
		child := RawFile{
			Path:     logical,
			FileName: baseName,
			MimeType: normalizeMime("", name),
			Data:     entryData,
		}

		if Classify(baseName, child.MimeType, entryData) == KindZip {
			if depth < s.limits.zipMaxDepth() {
				nestedArchive := archiveRoot + "/" + name
				nested, nestedNotes, _ := s.expand(entryData, nestedArchive, depth+1)
				leaves = append(leaves, nested...)
				notes = append(notes, nestedNotes...)
			} else {
				notes = append(notes, fmt.Sprintf("%s/%s: skipped (nested zip at depth limit)", archiveRoot, name))
			}
			continue
		}

		leaves = append(leaves, child)
	}

	return leaves, notes, nil
}

func cleanZipEntryPath(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "./")
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." {
		return ""
	}
	if strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return ""
	}
	return strings.Trim(cleaned, "/")
}

func rawFromSubmission(fileName, mimeType string, data []byte) RawFile {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "submission"
	}
	return RawFile{
		Path:     LogicalPath{RelativePath: name},
		FileName: path.Base(name),
		MimeType: normalizeMime(mimeType, name),
		Data:     data,
	}
}
