package models

type Skill struct {
	ID         int32  `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Category   int32  `json:"category"`
	Popularity int32  `json:"popularity"`
}
