package analysis

import (
	"encoding/json"

	"rubrical/internal/analysis/files"
	"rubrical/internal/analysis/prompt"
	"rubrical/internal/analysis/request"
)

func BuildProviderRequest(input Input, fileResult files.ProcessResult, maxSubmissionTextChars int) request.Request {
	fileContext := promptFileContextFrom(fileResult)
	system, user := prompt.Build(promptInputFrom(input, fileContext), maxSubmissionTextChars)

	var attachments []request.Attachment
	for _, file := range fileResult.Attachments {
		attachments = append(attachments, request.Attachment{
			Path:     file.Path.String(),
			MimeType: file.MimeType,
			Data:     file.Data,
			Delivery: request.DeliveryKind(file.Delivery),
		})
	}

	return request.Request{
		SystemPrompt: system,
		UserPrompt:   user,
		Attachments:  attachments,
	}
}

func promptFileContextFrom(result files.ProcessResult) prompt.FileContext {
	ctx := prompt.FileContext{
		SkippedNotes: append([]string(nil), result.SkippedNotes...),
	}
	for _, manifest := range result.Manifests {
		ctx.Manifests = append(ctx.Manifests, prompt.FileManifest{Tree: manifest.Tree})
	}
	for _, section := range result.InlineSections {
		ctx.InlineSections = append(ctx.InlineSections, prompt.FileInlineSection{
			Path:      section.Path.String(),
			Text:      section.Text,
			Extracted: section.Extracted,
		})
	}
	for _, file := range result.Attachments {
		ctx.AttachedFiles = append(ctx.AttachedFiles, prompt.AttachedFileIndex{
			Path:     file.Path.String(),
			MimeType: file.MimeType,
			Bytes:    len(file.Data),
		})
	}
	return ctx
}

func promptInputFrom(input Input, fileContext prompt.FileContext) prompt.Input {
	rows := make([]prompt.RubricRow, len(input.Rubric.Rows))
	for i, row := range input.Rubric.Rows {
		ratings := make([]prompt.RubricRating, len(row.Ratings))
		for j, rating := range row.Ratings {
			ratings[j] = prompt.RubricRating{
				Title:       rating.Title,
				Description: rating.Description,
				Points:      rating.Points,
			}
		}
		rows[i] = prompt.RubricRow{
			Criterion:                row.Criterion,
			CriterionLongDescription: row.CriterionLongDescription,
			Points:                   row.Points,
			Ratings:                  ratings,
		}
	}

	return prompt.Input{
		PageType:       input.PageType,
		Title:          input.Title,
		CourseName:     input.CourseName,
		Instructions:   input.Instructions,
		PointsPossible: input.PointsPossible,
		DraftMode:      input.DraftMode,
		DraftText:      input.DraftText,
		DraftURL:       input.DraftURL,
		Files:          fileContext,
		Rubric: prompt.Rubric{
			Header: append([]string(nil), input.Rubric.Header...),
			Rows:   rows,
		},
	}
}

type PromptLog struct {
	System string          `json:"system"`
	User   string          `json:"user"`
	Files  []PromptFileLog `json:"files,omitempty"`
}

type PromptFileLog struct {
	Path     string `json:"path"`
	MimeType string `json:"mimeType"`
	Delivery string `json:"delivery"`
	Bytes    int    `json:"bytes"`
}

func EncodePromptLog(req request.Request) ([]byte, error) {
	log := PromptLog{
		System: req.SystemPrompt,
		User:   req.UserPrompt,
	}
	for _, file := range req.Attachments {
		log.Files = append(log.Files, PromptFileLog{
			Path:     file.Path,
			MimeType: file.MimeType,
			Delivery: string(file.Delivery),
			Bytes:    len(file.Data),
		})
	}
	return json.Marshal(log)
}
