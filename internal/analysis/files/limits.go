package files

import (
	"rubrical/internal/config"
)

// Limits caps upload reads during zip extraction and total bytes sent to the model.
// Zip depth and per-archive uncompressed totals are hardcoded in config (not ENV).
type Limits struct {
	MaxUploadBytes int
	MaxTotalBytes  int
}

func (l Limits) withDefaults() Limits {
	if l.MaxUploadBytes <= 0 {
		l.MaxUploadBytes = config.DefaultDraftMaxUploadBytes
	}
	if l.MaxTotalBytes <= 0 {
		l.MaxTotalBytes = config.DefaultAnalysisMaxTotalBytes
	}
	return l
}

func (l Limits) zipMaxDepth() int {
	return config.DefaultAnalysisZipMaxDepth
}

func (l Limits) zipMaxUncompressedBytes() int {
	return config.DefaultAnalysisZipMaxUncompressedBytes
}

func LimitsFromConfig(maxUploadBytes, maxTotalBytes int) Limits {
	return Limits{
		MaxUploadBytes: maxUploadBytes,
		MaxTotalBytes:  maxTotalBytes,
	}.withDefaults()
}
