package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                         string
	Port                         int
	DatabaseURL                  string
	DataDir                      string
	SecretsEncryptionKey         string
	PublicURL                    string
	SessionTTL                   time.Duration
	GoogleOAuthClientID          string
	GoogleOAuthClientSecret      string
	ExtensionOrigins             []string
	EmailFrom                    string
	ResendAPIKey                 string
	SMTPHost                     string
	SMTPPort                     string
	SMTPUsername                 string
	SMTPPassword                 string
	EmailDevLog                  bool
	DraftMaxUploadBytes          int
	DraftMaxUploadSlots          int
	AnalysisMaxSubmissionTextChars int
	AnalysisMaxTotalBytes          int
	AIMaxRunsPerHour             int
	AIMaxRunsPerDay              int
	AIMinSecondsBetweenRuns      int
	AIEnforceRateLimits          bool
	StrictExtraction             bool
	AllowLocalURLFetch           bool
	PostDueDateRetention         time.Duration
	PostUploadRetention          time.Duration
}

func Load() (Config, error) {
	loadEnvFiles()

	databaseURL, err := loadDatabaseURL()
	if err != nil {
		return Config{}, err
	}

	// Listen host/port use homeserver app-key names in server.env (rubrical → RUBRICAL_*).
	cfg := Config{
		Host:                           strings.TrimSpace(envOrDefault("RUBRICAL_HOST", DefaultHost)),
		Port:                           envInt("RUBRICAL_PORT", DefaultPort),
		DatabaseURL:                    databaseURL,
		DataDir:                        envOrDefault("DATA_DIR", DefaultDataDir),
		SecretsEncryptionKey:           strings.TrimSpace(os.Getenv("SECRETS_ENCRYPTION_KEY")),
		PublicURL:                      strings.TrimRight(strings.TrimSpace(envOrDefault("PUBLIC_URL", DefaultPublicURL)), "/"),
		GoogleOAuthClientID:            strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID")),
		GoogleOAuthClientSecret:        strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")),
		ExtensionOrigins:               splitCSV(os.Getenv("EXTENSION_ORIGINS")),
		EmailFrom:                      envOrDefault("EMAIL_FROM", DefaultEmailFrom),
		ResendAPIKey:                   strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		SMTPHost:                       strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                       envOrDefault("SMTP_PORT", DefaultSMTPPort),
		SMTPUsername:                   strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:                   strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		EmailDevLog:                    envBool("EMAIL_DEV_LOG"),
		DraftMaxUploadBytes:            envInt("DRAFT_MAX_UPLOAD_BYTES", DefaultDraftMaxUploadBytes),
		DraftMaxUploadSlots:            envInt("DRAFT_MAX_UPLOAD_SLOTS", DefaultDraftMaxUploadSlots),
		AnalysisMaxSubmissionTextChars: envInt("ANALYSIS_MAX_SUBMISSION_TEXT_CHARS", DefaultAnalysisMaxSubmissionTextChars),
		AnalysisMaxTotalBytes:          envInt("ANALYSIS_MAX_TOTAL_BYTES", DefaultAnalysisMaxTotalBytes),
		AIMaxRunsPerHour:               envInt("AI_MAX_RUNS_PER_HOUR", DefaultAIMaxRunsPerHour),
		AIMaxRunsPerDay:                envInt("AI_MAX_RUNS_PER_DAY", DefaultAIMaxRunsPerDay),
		AIMinSecondsBetweenRuns:        envInt("AI_MIN_SECONDS_BETWEEN_RUNS", DefaultAIMinSecondsBetweenRuns),
		AIEnforceRateLimits:            envBool("AI_ENFORCE_RATE_LIMITS"),
		StrictExtraction:               envBool("STRICT_EXTRACTION"),
		AllowLocalURLFetch:             envBool("ALLOW_LOCAL_URL_FETCH"),
	}

	retention, err := envDuration("POST_DUE_DATE_RETENTION_TIME", DefaultPostDueDateRetention)
	if err != nil {
		return Config{}, fmt.Errorf("POST_DUE_DATE_RETENTION_TIME: %w", err)
	}
	cfg.PostDueDateRetention = retention

	uploadRetention, err := envDuration("POST_UPLOAD_RETENTION_TIME", DefaultPostUploadRetention)
	if err != nil {
		return Config{}, fmt.Errorf("POST_UPLOAD_RETENTION_TIME: %w", err)
	}
	cfg.PostUploadRetention = uploadRetention

	sessionTTL, err := envDuration("SESSION_TTL", DefaultSessionTTL)
	if err != nil {
		return Config{}, fmt.Errorf("SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = sessionTTL

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("RUBRICAL_PORT must be between 1 and 65535")
	}
	return cfg, nil
}

// loadDatabaseURL builds from POSTGRES_HOST/PORT (server.env) + POSTGRES_USER/PASSWORD/DB/SSLMODE (app env).
// Every piece is required — no defaults.
func loadDatabaseURL() (string, error) {
	host := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
	port := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
	user := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
	password := os.Getenv("POSTGRES_PASSWORD")
	db := strings.TrimSpace(os.Getenv("POSTGRES_DB"))
	sslmode := strings.TrimSpace(os.Getenv("POSTGRES_SSLMODE"))

	var missing []string
	if host == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if port == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if user == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if password == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if db == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if sslmode == "" {
		missing = append(missing, "POSTGRES_SSLMODE")
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   netHostPort(host, port),
		Path:   "/" + db,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func netHostPort(host, port string) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		return "[" + host + "]:" + port
	}
	return host + ":" + port
}

// ListenAddr is the net/http listen address from RUBRICAL_HOST + RUBRICAL_PORT.
// Empty host means all interfaces (":8787"), matching local WSL defaults.
func (c Config) ListenAddr() string {
	if c.Host == "" {
		return fmt.Sprintf(":%d", c.Port)
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return d, nil
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (c Config) CookieSecure() bool {
	return strings.HasPrefix(strings.ToLower(c.PublicURL), "https://")
}

func (c Config) GoogleOAuthRedirectURL() string {
	return strings.TrimRight(c.PublicURL, "/") + "/auth/google/callback"
}

func (c Config) AllowedOrigins() []string {
	origins := append([]string{}, c.ExtensionOrigins...)
	if c.PublicURL != "" {
		origins = append(origins, c.PublicURL)
	}
	origins = append(origins,
		fmt.Sprintf("http://localhost:%d", c.Port),
		fmt.Sprintf("http://127.0.0.1:%d", c.Port),
	)
	seen := make(map[string]struct{}, len(origins))
	out := make([]string, 0, len(origins))
	for _, origin := range origins {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		out = append(out, origin)
	}
	return out
}
