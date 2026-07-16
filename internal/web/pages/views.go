package pages

import "net/url"

const DashboardPath = "/"
const OnboardingPath = "/onboarding"

type AuthMode string

const (
	AuthModeLogin  AuthMode = "login"
	AuthModeSignup AuthMode = "signup"
	AuthModeForgot AuthMode = "forgot"
	AuthModeReset  AuthMode = "reset"
)

type AuthFormView struct {
	Mode           AuthMode
	Next           string
	GoogleEnabled  bool
	ErrorMessage   string
	SuccessMessage string
	ResetToken     string
}

func ParseAuthMode(raw string) AuthMode {
	switch raw {
	case "signup", "sign-up":
		return AuthModeSignup
	case "forgot", "forgot-password":
		return AuthModeForgot
	case "reset":
		return AuthModeReset
	default:
		return AuthModeLogin
	}
}

func AuthPageTitle(mode AuthMode) string {
	switch mode {
	case AuthModeSignup:
		return "Create account"
	case AuthModeForgot:
		return "Forgot password"
	case AuthModeReset:
		return "Reset password"
	default:
		return "Sign in"
	}
}

func AuthModeURL(mode AuthMode, next string) string {
	path := "/login"
	query := url.Values{}
	if mode != AuthModeLogin && mode != AuthModeReset {
		query.Set("mode", string(mode))
	}
	if next != "" {
		query.Set("next", next)
	}
	if encoded := query.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}

func googleAuthURL(next string) string {
	return authPathWithNext("/auth/google", next)
}

func authPathWithNext(path, next string) string {
	if next == "" {
		return path
	}
	return path + "?next=" + url.QueryEscape(next)
}
