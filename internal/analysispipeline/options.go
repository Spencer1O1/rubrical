package analysispipeline

import (
	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/config"
)

type Options struct {
	MaxSubmissionTextChars int
	MaxUploadBytes         int
	MaxTotalBytes          int
}

func (o Options) withDefaults() Options {
	if o.MaxSubmissionTextChars <= 0 {
		o.MaxSubmissionTextChars = config.DefaultAnalysisMaxSubmissionTextChars
	}
	if o.MaxUploadBytes <= 0 {
		o.MaxUploadBytes = config.DefaultDraftMaxUploadBytes
	}
	if o.MaxTotalBytes <= 0 {
		o.MaxTotalBytes = config.DefaultAnalysisMaxTotalBytes
	}
	return o
}

func OptionsFromConfig(cfg config.Config) Options {
	return Options{
		MaxSubmissionTextChars: cfg.AnalysisMaxSubmissionTextChars,
		MaxUploadBytes:         cfg.DraftMaxUploadBytes,
		MaxTotalBytes:          cfg.AnalysisMaxTotalBytes,
	}.withDefaults()
}

func (o Options) FileLimits() files.Limits {
	return files.LimitsFromConfig(o.MaxUploadBytes, o.MaxTotalBytes)
}
