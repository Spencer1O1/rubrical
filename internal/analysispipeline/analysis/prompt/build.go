package prompt

import (
	"strings"

	"rubrical/internal/analysispipeline/userprompt"
)

func Build(input Input, maxSubmissionTextChars int) (system string, user string) {
	system = BuildSystem(input.PageType)

	var b strings.Builder
	b.WriteString(BuildContext(input))
	b.WriteByte('\n')
	b.WriteString(userprompt.Instructions(input.Instructions))
	b.WriteByte('\n')
	b.WriteString(BuildRubric(input.Rubric))
	b.WriteByte('\n')
	b.WriteString(BuildSubmission(input, maxSubmissionTextChars))

	user = b.String()
	return system, user
}
