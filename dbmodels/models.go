package dbmodels

import (
	"database/sql"
	"time"
)

type Item struct {
	ID          int64 `gorm:"primary_key"`
	Completed   bool
	Description sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
}

func NullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
