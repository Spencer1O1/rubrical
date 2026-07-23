package analysispipeline

import (
	"encoding/json"
	"fmt"

	"rubrical/internal/llm"
)

type PromptLog struct {
	System string          `json:"system"`
	User   string          `json:"user"`
	Files  []PromptFileLog `json:"files,omitempty"`
}

// PipelinePromptLog stores every LLM pass sent for a run.
type PipelinePromptLog struct {
	Checkability PromptLog  `json:"checkability"`
	Analysis      *PromptLog `json:"analysis,omitempty"`
}

type PromptFileLog struct {
	Path     string `json:"path"`
	MimeType string `json:"mimeType"`
	Delivery string `json:"delivery"`
	Bytes    int    `json:"bytes"`
}

func ValidateLLMRequest(req llm.Request) error {
	if len(req.Schema) == 0 {
		return fmt.Errorf("json schema is required")
	}
	for _, file := range req.Attachments {
		switch file.Delivery {
		case llm.DeliveryPDF, llm.DeliveryImage, llm.DeliveryProviderFile:
		default:
			return fmt.Errorf("unsupported attachment delivery %q for %s", file.Delivery, file.Filename)
		}
	}
	return nil
}

func promptLogFromRequest(req llm.Request) PromptLog {
	log := PromptLog{
		System: req.SystemPrompt,
		User:   req.UserPrompt,
	}
	for _, file := range req.Attachments {
		log.Files = append(log.Files, PromptFileLog{
			Path:     file.Filename,
			MimeType: file.MimeType,
			Delivery: string(file.Delivery),
			Bytes:    len(file.Data),
		})
	}
	return log
}

// EncodePromptLog encodes a single request (tests / callers that only have one pass).
func EncodePromptLog(req llm.Request) ([]byte, error) {
	return json.Marshal(promptLogFromRequest(req))
}

// EncodePipelinePromptLog encodes pass 1 and optional pass 2 for raw_model_input.
func EncodePipelinePromptLog(pass1 llm.Request, pass2 *llm.Request) ([]byte, error) {
	out := PipelinePromptLog{Checkability: promptLogFromRequest(pass1)}
	if pass2 != nil {
		log := promptLogFromRequest(*pass2)
		out.Analysis = &log
	}
	return json.Marshal(out)
}
