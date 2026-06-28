package files

import (
	"bytes"
	"path/filepath"
	"strings"
)

type FileKind int

const (
	KindUnknown FileKind = iota
	KindText
	KindCode
	KindMarkdown
	KindJSON
	KindHTML
	KindXML
	KindPDF
	KindImage
	KindZip
	KindDocx
	KindSpreadsheet
	KindOffice
	KindLegacyDoc
	KindExecutable
	KindMedia
	KindArchiveUnsupported
)

func IsValidPDF(data []byte) bool {
	return len(data) >= 4 && bytes.HasPrefix(data, []byte("%PDF"))
}

func Classify(fileName, mimeType string, data []byte) FileKind {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "submission"
	}
	mime := normalizeMime(mimeType, name)
	ext := strings.ToLower(filepath.Ext(name))

	if IsValidPDF(data) {
		return KindPDF
	}
	if len(data) >= 2 && data[0] == 0x50 && data[1] == 0x4B {
		if ext == ".docx" || mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
			return KindDocx
		}
		if ext == ".zip" || mime == "application/zip" || mime == "application/x-zip-compressed" {
			return KindZip
		}
	}

	switch {
	case mime == "application/pdf" || ext == ".pdf":
		if IsValidPDF(data) {
			return KindPDF
		}
		return KindUnknown
	case strings.HasPrefix(mime, "image/"):
		return KindImage
	case ext == ".zip" || mime == "application/zip" || mime == "application/x-zip-compressed":
		return KindZip
	case isDocx(mime, ext):
		return KindDocx
	case ext == ".doc" || mime == "application/msword":
		return KindLegacyDoc
	case isExecutable(ext):
		return KindExecutable
	case isMedia(ext, mime):
		return KindMedia
	case isUnsupportedArchive(ext, mime):
		return KindArchiveUnsupported
	case isSpreadsheet(ext, mime):
		return KindSpreadsheet
	case isOffice(ext, mime):
		return KindOffice
	case ext == ".json" || mime == "application/json":
		return KindJSON
	case ext == ".html" || ext == ".htm" || mime == "text/html":
		return KindHTML
	case ext == ".xml" || mime == "application/xml" || mime == "text/xml":
		return KindXML
	case ext == ".md" || ext == ".markdown" || mime == "text/markdown" || mime == "application/markdown":
		return KindMarkdown
	case isCode(ext, mime):
		return KindCode
	case isTextMime(mime) || ext == ".txt" || ext == ".csv":
		return KindText
	default:
		if isMostlyText(data) {
			return KindText
		}
		return KindUnknown
	}
}

func SkipReason(kind FileKind) string {
	switch kind {
	case KindExecutable:
		return "unsupported binary (not analyzed)"
	case KindMedia:
		return "unsupported media (not analyzed)"
	case KindArchiveUnsupported:
		return "unsupported archive format (zip only)"
	case KindLegacyDoc:
		return "legacy Word format not supported (use docx or pdf)"
	case KindUnknown:
		return "could not decode as text"
	default:
		return "unsupported file type"
	}
}

func normalizeMime(mimeType, fileName string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	if mime != "" && mime != "application/octet-stream" {
		return mime
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
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
	case ".doc":
		return "application/msword"
	case ".zip":
		return "application/zip"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	default:
		return mimeType
	}
}

func isDocx(mime, ext string) bool {
	return mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" || ext == ".docx"
}

func isTextMime(mime string) bool {
	return strings.HasPrefix(mime, "text/")
}

func isCode(ext, mime string) bool {
	switch ext {
	case ".js", ".ts", ".jsx", ".tsx", ".py", ".go", ".java", ".c", ".cpp", ".cc", ".h", ".hpp",
		".rs", ".rb", ".php", ".swift", ".kt", ".kts", ".scala", ".sh", ".bash", ".zsh",
		".css", ".scss", ".sass", ".less", ".sql", ".r", ".m", ".cs", ".vue", ".svelte":
		return true
	}
	switch mime {
	case "application/javascript", "text/javascript", "application/typescript",
		"application/x-python", "text/x-python", "text/x-go", "text/x-java-source":
		return true
	}
	return false
}

func isSpreadsheet(ext, mime string) bool {
	switch ext {
	case ".csv", ".xls", ".xlsx", ".tsv":
		return true
	}
	switch mime {
	case "text/csv", "application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	}
	return false
}

func isOffice(ext, mime string) bool {
	switch ext {
	case ".ppt", ".pptx", ".odt", ".odp", ".rtf":
		return true
	}
	switch mime {
	case "application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.oasis.opendocument.text",
		"application/vnd.oasis.opendocument.presentation",
		"application/rtf", "text/rtf":
		return true
	}
	return false
}

func isExecutable(ext string) bool {
	switch ext {
	case ".exe", ".dll", ".so", ".dylib", ".bin", ".app", ".msi", ".deb", ".rpm":
		return true
	}
	return false
}

func isMedia(ext, mime string) bool {
	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma":
		return true
	}
	return strings.HasPrefix(mime, "video/") || strings.HasPrefix(mime, "audio/")
}

func isUnsupportedArchive(ext, mime string) bool {
	switch ext {
	case ".rar", ".7z", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".tar.gz":
		return true
	}
	switch mime {
	case "application/x-rar-compressed", "application/x-7z-compressed",
		"application/x-tar", "application/gzip", "application/x-bzip2":
		return true
	}
	return false
}

func isMostlyText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	sample := data
	if len(sample) > 8192 {
		sample = sample[:8192]
	}
	for _, b := range sample {
		if b == 0 {
			return false
		}
	}
	return true
}
