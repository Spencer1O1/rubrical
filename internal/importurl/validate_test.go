package importurl

import "testing"

func TestValidateSourceURL(t *testing.T) {
	got, err := ValidateSourceURL("https://usu.instructure.com/courses/1/assignments/2?submitting=1")
	if err != nil {
		t.Fatalf("ValidateSourceURL: %v", err)
	}
	if got != "https://usu.instructure.com/courses/1/assignments/2" {
		t.Fatalf("got %q", got)
	}

	if _, err := ValidateSourceURL("https://example.com/courses/1/assignments/2"); err == nil {
		t.Fatal("expected non-canvas host to fail")
	}

	if _, err := ValidateSourceURL("https://school.instructure.com/courses/1/modules/2"); err == nil {
		t.Fatal("expected non-assignment path to fail")
	}
}
