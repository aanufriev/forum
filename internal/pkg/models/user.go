package models

//easyjson:json
type User struct {
	Nickname string  `json:"nickname"`
	Fullname *string `json:"fullname"`
	Email    *string `json:"email"`
	About    string  `json:"about"`
}
