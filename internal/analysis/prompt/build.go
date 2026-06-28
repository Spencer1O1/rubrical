package prompt

import "strings"

func Build(input Input, maxPromptDraftChars int) (system string, user string) {
	maxPromptDraftChars = normalizeMaxDraftChars(maxPromptDraftChars)
	system = BuildSystem()

	var b strings.Builder
	b.WriteString(BuildContext(input))
	b.WriteString(BuildInstructions(input.Instructions, maxPromptDraftChars/2))
	b.WriteString(BuildRubric(input.Rubric))
	b.WriteString(BuildSubmission(input, maxPromptDraftChars))

	user = b.String()
	return system, user
}
