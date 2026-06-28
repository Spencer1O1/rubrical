package importmeta

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	pointsNumberRE = regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	atTimeRE       = regexp.MustCompile(`(?i)\s+at\s+`)
)

var dueDateLayouts = []string{
	"Monday, January 2, 2006",
	"Monday, Jan 2, 2006",
	"January 2, 2006",
	"Jan 2, 2006",
	"January 2",
	"Jan 2",
	"1/2/2006",
	"1/2",
	"2006-01-02",
}

var combinedDueLayouts = []string{
	"Mon Jan 2, 2006 3:04pm",
	"Mon Jan 2, 2006 3:04 PM",
	"Monday, January 2, 2006 3:04 PM",
	"Monday, January 2, 2006 3:04pm",
}

var dueTimeLayouts = []string{
	"3:04pm",
	"3:04 PM",
	"15:04",
	"3pm",
	"3 PM",
}

// ParseDueAtISO parses Canvas ENV due_at timestamps.
func ParseDueAtISO(iso string) (time.Time, bool) {
	iso = strings.TrimSpace(iso)
	if iso == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, iso); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// ParseDueAt extracts a due timestamp from Canvas label text such as "Due Jun 26 at 11:59pm".
// ref supplies the year for dates without one and resolves "Today"/"Tomorrow".
func ParseDueAt(text string, ref time.Time) (time.Time, bool) {
	cleaned := stripDuePrefix(text)
	if cleaned == "" || isNoDueDate(cleaned) {
		return time.Time{}, false
	}

	loc := ref.Location()

	for _, layout := range combinedDueLayouts {
		if t, err := time.ParseInLocation(layout, cleaned, loc); err == nil {
			return t, true
		}
	}

	datePart, timePart := splitDateAndTime(cleaned)

	switch strings.ToLower(strings.TrimSpace(datePart)) {
	case "today":
		return combineDateTime(truncateToDate(ref, loc), timePart, loc)
	case "tomorrow":
		return combineDateTime(truncateToDate(ref.Add(24*time.Hour), loc), timePart, loc)
	}

	parsedDate, ok := parseDueDatePart(datePart, ref, loc)
	if !ok {
		return time.Time{}, false
	}

	return combineDateTime(parsedDate, timePart, loc)
}

// ParsePointsPossible extracts numeric points from Canvas label text such as "25 pts".
func ParsePointsPossible(text string) (float64, bool) {
	cleaned := strings.TrimSpace(text)
	if cleaned == "" || cleaned == "—" || cleaned == "-" {
		return 0, false
	}

	lower := strings.ToLower(cleaned)
	if strings.Contains(lower, "no points") {
		return 0, false
	}

	match := pointsNumberRE.FindStringSubmatch(cleaned)
	if len(match) < 2 {
		return 0, false
	}

	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil || value < 0 {
		return 0, false
	}

	return value, true
}

func stripDuePrefix(text string) string {
	cleaned := strings.TrimSpace(text)
	for strings.HasPrefix(strings.ToLower(cleaned), "due") {
		cleaned = strings.TrimSpace(cleaned[3:])
		cleaned = strings.TrimLeft(cleaned, " :")
	}
	return cleaned
}

func isNoDueDate(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(lower, "no due date")
}

func splitDateAndTime(text string) (string, string) {
	parts := atTimeRE.Split(text, 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func truncateToDate(t time.Time, loc *time.Location) time.Time {
	if loc != nil {
		t = t.In(loc)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func parseDueDatePart(datePart string, ref time.Time, loc *time.Location) (time.Time, bool) {
	datePart = strings.TrimSpace(datePart)
	if datePart == "" {
		return time.Time{}, false
	}

	for _, layout := range dueDateLayouts {
		if t, err := time.ParseInLocation(layout, datePart, loc); err == nil {
			if !strings.Contains(layout, "2006") {
				_, m, d := t.Date()
				t = time.Date(ref.Year(), m, d, 0, 0, 0, 0, loc)
			}
			return t, true
		}
	}

	return time.Time{}, false
}

func combineDateTime(date time.Time, timePart string, loc *time.Location) (time.Time, bool) {
	if strings.TrimSpace(timePart) == "" {
		end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, loc)
		return end, true
	}

	for _, layout := range dueTimeLayouts {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(timePart), loc); err == nil {
			combined := time.Date(
				date.Year(), date.Month(), date.Day(),
				parsed.Hour(), parsed.Minute(), parsed.Second(), 0, loc,
			)
			return combined, true
		}
	}

	return time.Time{}, false
}
