package criterion

import (
	"fmt"
	"strings"
	"unicode"
)

// Ref is one rubric criterion for prompts/schemas: stable id, display name, optional long description.
type Ref struct {
	ID          string
	Name        string
	Description string
}

// Index builds unique slug ids for rubric criterion names (rubric order).
func Index(names []string) []Ref {
	bases := make([]string, len(names))
	for i, name := range names {
		bases[i] = Slugify(name)
		if bases[i] == "" {
			bases[i] = "criterion"
		}
	}
	ids := Unique(bases)
	out := make([]Ref, len(names))
	for i, name := range names {
		out[i] = Ref{ID: ids[i], Name: name}
	}
	return out
}

// Lookup returns the ref for id, or false.
func Lookup(refs []Ref, id string) (Ref, bool) {
	for _, r := range refs {
		if r.ID == id {
			return r, true
		}
	}
	return Ref{}, false
}

// Slugify lowercases and replaces non-alphanumeric runs with a single '-'.
func Slugify(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range strings.TrimSpace(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
		default:
			if b.Len() > 0 && !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// Unique makes ids distinct by appending -2, -3, … on collision.
func Unique(bases []string) []string {
	out := make([]string, len(bases))
	used := make(map[string]int, len(bases))
	for i, base := range bases {
		if base == "" {
			base = "id"
		}
		n := used[base]
		used[base] = n + 1
		if n == 0 {
			out[i] = base
			continue
		}
		candidate := fmt.Sprintf("%s-%d", base, n+1)
		for used[candidate] > 0 {
			n++
			used[base] = n + 1
			candidate = fmt.Sprintf("%s-%d", base, n+1)
		}
		used[candidate] = 1
		out[i] = candidate
	}
	return out
}

// RatingID is the closed id for band i within a criterion (sorted low→high points).
func RatingID(bandIndex int) string {
	return fmt.Sprintf("r%d", bandIndex)
}
