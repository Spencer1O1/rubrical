package analysis

import (
	"rubrical/internal/analysis/prompt"
	"rubrical/internal/config"
)

const defaultPromptMaxDraftChars = prompt.DefaultMaxDraftChars

type Options struct {
	PromptMaxDraftChars int
	MaxFileBytes        int
}

func (o Options) withDefaults() Options {
	if o.PromptMaxDraftChars <= 0 {
		o.PromptMaxDraftChars = defaultPromptMaxDraftChars
	}
	if o.MaxFileBytes <= 0 {
		o.MaxFileBytes = config.DefaultDraftMaxFileBytes
	}
	return o
}

func OptionsFromConfig(promptMaxDraftChars, draftMaxFileBytes int) Options {
	return Options{
		PromptMaxDraftChars: promptMaxDraftChars,
		MaxFileBytes:        draftMaxFileBytes,
	}.withDefaults()
}
