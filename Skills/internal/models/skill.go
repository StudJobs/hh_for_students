package models

import "time"

type Skill struct {
	ID         int32     `db:"id"`
	Slug       string    `db:"slug"`
	Name       string    `db:"name"`
	Category   int32     `db:"category"`
	Popularity int32     `db:"popularity"`
	CreatedAt  time.Time `db:"created_at"`
}
