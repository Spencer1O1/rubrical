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
	Files          []File
	Rubric         Rubric
}

type File struct {
	FileName string
	MimeType string
	Data     []byte
}

type Rubric struct {
	Header []string
	Rows   []RubricRow
}

type RubricRow struct {
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
