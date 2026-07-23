package analysispipeline

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func criterionStatusLabel(status string) string {
	switch status {
	case "met":
		return "This criterion appears to be met."
	case "partially_met":
		return "This criterion is partially met."
	case "not_met":
		return "This criterion does not appear to be met."
	case "not_checkable":
		return "This criterion can’t be checked from this draft."
	default:
		return ""
	}
}

func errorsIsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func nullIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func numericToFloatPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}
