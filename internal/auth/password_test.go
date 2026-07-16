package auth

import "testing"

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "correct horse battery") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Fatal("expected wrong password to fail")
	}
}

func TestHashPasswordRejectsWeakPassword(t *testing.T) {
	if _, err := HashPassword("short"); err != ErrWeakPassword {
		t.Fatalf("err=%v", err)
	}
}

func TestSanitizeNextPath(t *testing.T) {
	if SanitizeNextPath("/assignments/1") != "/assignments/1" {
		t.Fatal("expected valid next path")
	}
	if SanitizeNextPath("/assignments/1?embed=1") != "/assignments/1?embed=1" {
		t.Fatal("expected embed query preserved")
	}
	if SanitizeNextPath("//evil") != "" {
		t.Fatal("expected protocol-relative path rejected")
	}
	if SanitizeNextPath("https://evil") != "" {
		t.Fatal("expected absolute URL rejected")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if NormalizeEmail("  User@Example.COM ") != "user@example.com" {
		t.Fatal("expected normalized email")
	}
}

func TestPasswordLoginMessage(t *testing.T) {
	if PasswordLoginMessage(ErrNoPassword) == "Invalid email or password." {
		t.Fatal("expected Google/forgot-password guidance for ErrNoPassword")
	}
	if PasswordLoginMessage(ErrInvalidCredentials) != "Invalid email or password." {
		t.Fatal("expected generic credentials message")
	}
}

func TestSignupMessage(t *testing.T) {
	msg := SignupMessage(ErrEmailTaken)
	if msg == ErrEmailTaken.Error() {
		t.Fatal("expected guidance beyond raw ErrEmailTaken")
	}
}
