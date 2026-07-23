package prompt

type Input struct {
	PageType       string
	Title          string
	CourseName     string
	Instructions   string
	PointsPossible *float64
	DraftMode      string
	DraftText      string
	DraftURL       string
	Files          FileContext
	Rubric         Rubric
}

type Rubric struct {
	Header []string     `json:"header"`
	Rows   []RubricRow  `json:"rows"`
}

type RubricRow struct {
	ID                       string         `json:"id"`
	Criterion                string         `json:"criterion"`
	CriterionLongDescription string         `json:"criterionLongDescription,omitempty"`
	Points                   string         `json:"points,omitempty"`
	Ratings                  []RubricRating `json:"ratings"`
}

type RubricRating struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Points      string `json:"points,omitempty"`
}
