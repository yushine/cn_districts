package models

import "time"

type District struct {
	ID        int
	CountryID int
	Name      string
	Level     int
	UpID      int
	Code      string
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
}
