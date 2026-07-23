package config

import "time"

// Server defaults (ENV: RUBRICAL_HOST/PORT from homeserver app key, POSTGRES_*, DATA_DIR, POST_*_RETENTION_TIME).
// Dev flags without constants here (ENV only, see config.Load): STRICT_EXTRACTION, ALLOW_LOCAL_URL_FETCH.
const (
	DefaultHost                 = "" // empty host → listen on all interfaces (:PORT)
	DefaultPort                 = 8787
	DefaultDataDir              = "./data"
	DefaultPostDueDateRetention = 168 * time.Hour
	DefaultPostUploadRetention  = 720 * time.Hour
	DefaultPublicURL            = "http://localhost:8787"
	DefaultSessionTTL           = 30 * 24 * time.Hour
	DefaultEmailFrom            = "Rubrical <rubrical@spencerls.dev>"
	DefaultSMTPPort             = "587"
)

// Draft upload limits (ENV: DRAFT_MAX_UPLOAD_BYTES, DRAFT_MAX_UPLOAD_SLOTS).
// Canvas file-upload submissions allow up to 5 GiB per file; media recordings up to 500 MiB.
// Rubrical stores bytes locally — default matches Canvas media / typical assignment size, not 5 GiB.
const (
	DefaultDraftMaxUploadBytes = 500 << 20
	DefaultDraftMaxUploadSlots = 20
)

// Analysis limits (ENV: ANALYSIS_MAX_SUBMISSION_TEXT_CHARS, ANALYSIS_MAX_TOTAL_BYTES).
// Submission text = typed draft + inline extracted file text (one shared char pool).
// File bytes = binary payloads sent as native attachments (separate from text pool).
const (
	DefaultAnalysisMaxSubmissionTextChars = 120_000
	DefaultAnalysisMaxManifestChars       = 32_000
	DefaultAnalysisMaxTotalBytes          = 64 << 20
)

// Zip extraction safety (hardcoded — not ENV; change in code).
const (
	DefaultAnalysisZipMaxDepth           = 2
	DefaultAnalysisZipMaxUncompressedBytes = 128 << 20
)

// Rate limits (ENV: AI_ENFORCE_RATE_LIMITS, AI_MAX_RUNS_*). Zero = unlimited.
const (
	DefaultAIMaxRunsPerHour        = 0
	DefaultAIMaxRunsPerDay         = 0
	DefaultAIMinSecondsBetweenRuns = 0
)

// Per-user AI defaults (stored in user_ai_settings when empty; not ENV).
const (
	DefaultAIProvider     = "openai"
	DefaultOpenAIModel    = "gpt-4o-mini"
	DefaultAnthropicModel = "claude-sonnet-5"
)

// Import payload caps (hardcoded — protect server from huge JSON; not ENV).
const (
	DefaultImportMaxBodyBytes     = 8 << 20
	DefaultMaxTitleRunes          = 500
	DefaultMaxCourseNameRunes     = 300
	DefaultMaxMetadataFieldRunes  = 500
	DefaultMaxVisibleTextBytes    = 512 << 10
	DefaultMaxInstructionsBytes   = 512 << 10
	DefaultMaxDraftTextBytes      = 512 << 10
	DefaultMaxRubricRows          = 100
	DefaultMaxRubricHeaderColumns = 20
	DefaultMaxRatingsPerRow       = 50
	DefaultMaxRubricFieldRunes    = 4000
	DefaultMaxDraftFileNameRunes  = 255
)

// Provider client defaults (hardcoded — API wiring; not ENV or per-user).
const (
	DefaultOpenAIBaseURL      = "https://api.openai.com/v1"
	DefaultAnthropicBaseURL   = "https://api.anthropic.com/v1"
	DefaultProviderTimeout    = 120 * time.Second
	DefaultURLFetchTimeout    = 15 * time.Second
	DefaultOpenAITemperature  = 0.2
	DefaultAnthropicMaxTokens = 8192
)
