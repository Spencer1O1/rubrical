package handlers

import (
	"rubrical/internal/db"
)

type Handlers struct {
	db               *db.DB
	userID           int64
	strictExtraction bool
}

func New(database *db.DB, userID int64, strictExtraction bool) *Handlers {
	return &Handlers{
		db:               database,
		userID:           userID,
		strictExtraction: strictExtraction,
	}
}
