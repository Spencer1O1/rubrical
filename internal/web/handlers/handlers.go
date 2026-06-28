package handlers

import (
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
)

type Handlers struct {
	db               *db.DB
	files            *draftfiles.Store
	userID           int64
	strictExtraction bool
}

func New(database *db.DB, files *draftfiles.Store, userID int64, strictExtraction bool) *Handlers {
	return &Handlers{
		db:               database,
		files:            files,
		userID:           userID,
		strictExtraction: strictExtraction,
	}
}
