package importpayload

import "rubrical/internal/config"

const (
	defaultImportMaxBodyBytes   = 8 << 20
	defaultMaxTitleRunes          = 500
	defaultMaxCourseNameRunes     = 300
	defaultMaxMetadataFieldRunes  = 500
	defaultMaxVisibleTextBytes    = 512 << 10
	defaultMaxInstructionsBytes   = 512 << 10
	defaultMaxDraftTextBytes      = 512 << 10
	defaultMaxRubricRows          = 100
	defaultMaxRubricHeaderColumns = 20
	defaultMaxRatingsPerRow       = 50
	defaultMaxRubricFieldRunes    = 4000
	defaultMaxDraftFileNameRunes  = 255
	defaultMaxDraftFiles          = 20
)

// Limits caps extension import payloads (POST /imports).
type Limits struct {
	MaxBodyBytes          int
	MaxTitleRunes         int
	MaxCourseNameRunes    int
	MaxMetadataFieldRunes int
	MaxVisibleTextBytes   int
	MaxInstructionsBytes  int
	MaxDraftTextBytes     int
	MaxRubricRows         int
	MaxRubricHeaderColumns int
	MaxRatingsPerRow      int
	MaxRubricFieldRunes   int
	MaxDraftFileNameRunes int
	MaxFileBytes          int
	MaxFiles              int
}

func DefaultLimits() Limits {
	return Limits{
		MaxBodyBytes:           defaultImportMaxBodyBytes,
		MaxTitleRunes:          defaultMaxTitleRunes,
		MaxCourseNameRunes:     defaultMaxCourseNameRunes,
		MaxMetadataFieldRunes:  defaultMaxMetadataFieldRunes,
		MaxVisibleTextBytes:    defaultMaxVisibleTextBytes,
		MaxInstructionsBytes:   defaultMaxInstructionsBytes,
		MaxDraftTextBytes:      defaultMaxDraftTextBytes,
		MaxRubricRows:          defaultMaxRubricRows,
		MaxRubricHeaderColumns: defaultMaxRubricHeaderColumns,
		MaxRatingsPerRow:       defaultMaxRatingsPerRow,
		MaxRubricFieldRunes:    defaultMaxRubricFieldRunes,
		MaxDraftFileNameRunes:  defaultMaxDraftFileNameRunes,
		MaxFileBytes:           config.DefaultDraftMaxFileBytes,
		MaxFiles:               defaultMaxDraftFiles,
	}
}

func LimitsFromConfig(draftMaxFileBytes, draftMaxFiles int) Limits {
	return Limits{
		MaxFileBytes: draftMaxFileBytes,
		MaxFiles:     draftMaxFiles,
	}.WithDefaults()
}

func (l Limits) WithDefaults() Limits {
	d := DefaultLimits()
	if l.MaxBodyBytes <= 0 {
		l.MaxBodyBytes = d.MaxBodyBytes
	}
	if l.MaxTitleRunes <= 0 {
		l.MaxTitleRunes = d.MaxTitleRunes
	}
	if l.MaxCourseNameRunes <= 0 {
		l.MaxCourseNameRunes = d.MaxCourseNameRunes
	}
	if l.MaxMetadataFieldRunes <= 0 {
		l.MaxMetadataFieldRunes = d.MaxMetadataFieldRunes
	}
	if l.MaxVisibleTextBytes <= 0 {
		l.MaxVisibleTextBytes = d.MaxVisibleTextBytes
	}
	if l.MaxInstructionsBytes <= 0 {
		l.MaxInstructionsBytes = d.MaxInstructionsBytes
	}
	if l.MaxDraftTextBytes <= 0 {
		l.MaxDraftTextBytes = d.MaxDraftTextBytes
	}
	if l.MaxRubricRows <= 0 {
		l.MaxRubricRows = d.MaxRubricRows
	}
	if l.MaxRubricHeaderColumns <= 0 {
		l.MaxRubricHeaderColumns = d.MaxRubricHeaderColumns
	}
	if l.MaxRatingsPerRow <= 0 {
		l.MaxRatingsPerRow = d.MaxRatingsPerRow
	}
	if l.MaxRubricFieldRunes <= 0 {
		l.MaxRubricFieldRunes = d.MaxRubricFieldRunes
	}
	if l.MaxDraftFileNameRunes <= 0 {
		l.MaxDraftFileNameRunes = d.MaxDraftFileNameRunes
	}
	if l.MaxFileBytes <= 0 {
		l.MaxFileBytes = d.MaxFileBytes
	}
	if l.MaxFiles <= 0 {
		l.MaxFiles = d.MaxFiles
	}
	return l
}
