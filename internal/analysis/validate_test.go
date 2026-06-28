package analysis

import (
	"encoding/json"
	"testing"

	"rubrical/internal/analysis/prompt"
)

func TestEncodePromptLog(t *testing.T) {
	data, err := EncodePromptLog(BuildProviderRequest(Input{
		Files: []SubmissionFile{{
			FileName: "essay.pdf",
			MimeType: "application/pdf",
			Data:     []byte("abc"),
		}},
	}, prompt.DefaultMaxDraftChars))
	if err != nil {
		t.Fatal(err)
	}
	var decoded PromptLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Files) != 1 || decoded.Files[0].Bytes != 3 {
		t.Fatalf("unexpected log: %+v", decoded)
	}
}
