package prompt

import "strings"

func Build(input Input, maxSubmissionTextChars int) (system string, user string) {
	system = BuildSystem()

	var b strings.Builder
	b.WriteString(BuildContext(input))
	b.WriteString(BuildInstructions(input.Instructions))
	b.WriteString(BuildRubric(input.Rubric))
	b.WriteString(BuildSubmission(input, maxSubmissionTextChars))

	user = b.String()
	return system, user
}
