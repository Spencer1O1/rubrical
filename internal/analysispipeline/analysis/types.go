package analysis

// RubricContext is the assignment rubric used by the analysis (pass 2) scoring path.
type RubricContext struct {
	Header []string
	Rows   []RubricRow
}

type RubricRow struct {
	ID                       string // stable slug for LLM schemas; set via AssignCriterionIDs
	Criterion                string
	CriterionLongDescription string
	Points                   string
	Ratings                  []RubricRating
}

type RubricRating struct {
	Title       string
	Description string
	Points      string
}

// DraftInput is the assignment + draft context for building an analysis LLM request.
type DraftInput struct {
	PageType       string
	Title          string
	CourseName     string
	Instructions   string
	PointsPossible *float64
	DraftMode      string
	DraftText      string
	DraftURL       string
	Rubric         RubricContext
}
