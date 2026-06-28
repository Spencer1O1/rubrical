package prompt

const DefaultMaxDraftChars = 120_000

func normalizeMaxDraftChars(max int) int {
	if max <= 0 {
		return DefaultMaxDraftChars
	}
	return max
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "\n…[truncated]"
}
