package prompt

import (
	"fmt"
	"strings"
)

func BuildContext(input Input) string {
	var b strings.Builder

	titleLabel := "Assignment title"
	if input.PageType == "discussion" {
		titleLabel = "Discussion title"
	}
	fmt.Fprintf(&b, "%s: %s\n", titleLabel, strings.TrimSpace(input.Title))

	if course := strings.TrimSpace(input.CourseName); course != "" {
		fmt.Fprintf(&b, "Course: %s\n", course)
	}

	if input.PointsPossible != nil {
		fmt.Fprintf(&b, "Points possible: %.2f\n", *input.PointsPossible)
	}
	return b.String()
}
