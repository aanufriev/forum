package models

//easyjson:json
type User struct {
	Nickname string  `json:"nickname"`
	Fullname *string `json:"fullname,omitempty"`
	Email    *string `json:"email,omitempty"`
	About    *string `json:"about,omitempty"`
}
