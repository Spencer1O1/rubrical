package analysis

import (
	"strings"

	"rubrical/internal/analysispipeline/analysis/prompt"
	"rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/analysispipeline/criterion"
	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/llm"
)

func BuildAnalysisRequest(input DraftInput, fileResult files.ProcessResult, maxSubmissionTextChars int, providerName string, rubric RubricContext) llm.Request {
	_ = ensureRubricIDs(&rubric)
	fileContext := promptFileContextFrom(fileResult)
	promptInput := promptInputFrom(input, fileContext, rubric)
	system, user := prompt.Build(promptInput, maxSubmissionTextChars)

	var attachments []llm.Attachment
	for _, file := range fileResult.Attachments {
		attachments = append(attachments, llm.Attachment{
			Filename: file.Path.String(),
			MimeType: file.MimeType,
			Data:     file.Data,
			Delivery: file.Delivery,
		})
	}

	criteria := rubricCriterionSpecs(rubric)
	return llm.Request{
		SystemPrompt: system,
		UserPrompt:   user,
		Attachments:  attachments,
		SchemaName:   "rubric_analysis",
		Schema:       analysisSchemaForProvider(providerName, criteria),
	}
}

func promptRubricFrom(rubric RubricContext) prompt.Rubric {
	rows := make([]prompt.RubricRow, len(rubric.Rows))
	for i, row := range rubric.Rows {
		bands := parseRatingBands(row.Ratings)
		ratings := make([]prompt.RubricRating, len(bands))
		for j, band := range bands {
			ratings[j] = prompt.RubricRating{
				ID:          criterion.RatingID(j),
				Title:       band.rating.Title,
				Description: band.rating.Description,
				Points:      band.rating.Points,
			}
		}
		rows[i] = prompt.RubricRow{
			ID:                       row.ID,
			Criterion:                row.Criterion,
			CriterionLongDescription: row.CriterionLongDescription,
			Points:                   row.Points,
			Ratings:                  ratings,
		}
	}
	return prompt.Rubric{
		Header: append([]string(nil), rubric.Header...),
		Rows:   rows,
	}
}

func analysisSchemaForProvider(providerName string, criteria []schema.CriterionSpec) map[string]any {
	switch strings.ToLower(strings.TrimSpace(providerName)) {
	case llm.ProviderAnthropic:
		return schema.JSONSchemaForAnthropic(criteria)
	default:
		return schema.JSONSchemaForOpenAI(criteria)
	}
}

func rubricCriterionSpecs(rubric RubricContext) []schema.CriterionSpec {
	if len(rubric.Rows) == 0 {
		return nil
	}
	specs := make([]schema.CriterionSpec, len(rubric.Rows))
	for i, row := range rubric.Rows {
		specs[i] = schema.CriterionSpec{
			ID:        row.ID,
			Name:      row.Criterion,
			RatingIDs: ratingIDsForSchema(row),
		}
	}
	return specs
}

func ratingIDsForSchema(row RubricRow) []string {
	bands := parseRatingBands(row.Ratings)
	if len(bands) == 0 {
		return []string{""}
	}
	ids := make([]string, len(bands))
	for i := range bands {
		ids[i] = criterion.RatingID(i)
	}
	return ids
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

func promptInputFrom(input DraftInput, fileContext prompt.FileContext, rubric RubricContext) prompt.Input {
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
		Rubric:         promptRubricFrom(rubric),
	}
}
