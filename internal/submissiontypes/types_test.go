package submissiontypes

import "testing"

func TestAttachmentsAllowed(t *testing.T) {
	if !AttachmentsAllowed([]string{"online_text_entry", "online_upload"}) {
		t.Fatal("expected online_upload to allow attachments")
	}
	if AttachmentsAllowed([]string{"online_text_entry"}) {
		t.Fatal("expected text-only to disallow attachments")
	}
}
