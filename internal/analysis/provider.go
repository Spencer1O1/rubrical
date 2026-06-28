package analysis

import (
	"encoding/json"
	"strings"

	"rubrical/internal/analysis/prompt"
	"rubrical/internal/analysis/request"
)

func BuildProviderRequest(input Input, maxPromptDraftChars int) request.Request {
	system, user := prompt.Build(promptInputFrom(input), maxPromptDraftChars)

	var attachments []request.Attachment
	for _, file := range input.Files {
		attachments = append(attachments, request.Attachment{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Data:     file.Data,
			Kind:     attachmentKind(file.MimeType, file.FileName),
		})
	}

	return request.Request{
		SystemPrompt: system,
		UserPrompt:   user,
		Attachments:  attachments,
	}
}

func promptInputFrom(input Input) prompt.Input {
	files := make([]prompt.File, len(input.Files))
	for i, file := range input.Files {
		files[i] = prompt.File{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Data:     file.Data,
		}
	}

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
		Files:          files,
		Rubric: prompt.Rubric{
			Header: append([]string(nil), input.Rubric.Header...),
			Rows:   rows,
		},
	}
}

func attachmentKind(mimeType, fileName string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case mime == "application/pdf" || strings.HasSuffix(strings.ToLower(fileName), ".pdf"):
		return "pdf"
	case strings.HasPrefix(mime, "image/"):
		return "image"
	default:
		return "text"
	}
}

type PromptLog struct {
	System string          `json:"system"`
	User   string          `json:"user"`
	Files  []PromptFileLog `json:"files,omitempty"`
}

type PromptFileLog struct {
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Kind     string `json:"kind"`
	Bytes    int    `json:"bytes"`
}

func EncodePromptLog(req request.Request) ([]byte, error) {
	log := PromptLog{
		System: req.SystemPrompt,
		User:   req.UserPrompt,
	}
	for _, file := range req.Attachments {
		log.Files = append(log.Files, PromptFileLog{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Kind:     file.Kind,
			Bytes:    len(file.Data),
		})
	}
	return json.Marshal(log)
}
