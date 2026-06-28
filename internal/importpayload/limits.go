package importpayload

import "rubrical/internal/config"

// Limits caps extension import payloads (POST /imports).
type Limits struct {
	MaxBodyBytes           int
	MaxTitleRunes          int
	MaxCourseNameRunes     int
	MaxMetadataFieldRunes  int
	MaxVisibleTextBytes    int
	MaxInstructionsBytes   int
	MaxDraftTextBytes      int
	MaxRubricRows          int
	MaxRubricHeaderColumns int
	MaxRatingsPerRow       int
	MaxRubricFieldRunes    int
	MaxDraftFileNameRunes  int
	MaxUploadBytes         int
	MaxUploadSlots         int
}

func DefaultLimits() Limits {
	return Limits{
		MaxBodyBytes:           config.DefaultImportMaxBodyBytes,
		MaxTitleRunes:          config.DefaultMaxTitleRunes,
		MaxCourseNameRunes:     config.DefaultMaxCourseNameRunes,
		MaxMetadataFieldRunes:  config.DefaultMaxMetadataFieldRunes,
		MaxVisibleTextBytes:    config.DefaultMaxVisibleTextBytes,
		MaxInstructionsBytes:   config.DefaultMaxInstructionsBytes,
		MaxDraftTextBytes:      config.DefaultMaxDraftTextBytes,
		MaxRubricRows:          config.DefaultMaxRubricRows,
		MaxRubricHeaderColumns: config.DefaultMaxRubricHeaderColumns,
		MaxRatingsPerRow:       config.DefaultMaxRatingsPerRow,
		MaxRubricFieldRunes:    config.DefaultMaxRubricFieldRunes,
		MaxDraftFileNameRunes:  config.DefaultMaxDraftFileNameRunes,
		MaxUploadBytes:         config.DefaultDraftMaxUploadBytes,
		MaxUploadSlots:         config.DefaultDraftMaxUploadSlots,
	}
}

func LimitsFromConfig(draftMaxUploadBytes, draftMaxUploadSlots int) Limits {
	return Limits{
		MaxUploadBytes: draftMaxUploadBytes,
		MaxUploadSlots: draftMaxUploadSlots,
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
	if l.MaxUploadBytes <= 0 {
		l.MaxUploadBytes = d.MaxUploadBytes
	}
	if l.MaxUploadSlots <= 0 {
		l.MaxUploadSlots = d.MaxUploadSlots
	}
	return l
}
