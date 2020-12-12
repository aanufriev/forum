package models

//easyjson:json
type User struct {
	Nickname string  `json:"nickname"`
	Fullname *string `json:"fullname,omitempty"`
	Email    *string `json:"email,omitempty"`
	About    *string `json:"about,omitempty"`
}

//easyjson:json
type Vote struct {
	Nickname string `json:"nickname"`
	Voice    int    `json:"voice"`
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
}
