package models

//easyjson:json
type Forum struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}

//easyjson:json
type Thread struct {
	Author  string  `json:"author"`
	Created *string `json:"created"`
	Forum   string  `json:"forum"`
	ID      int     `json:"id"`
	Message string  `json:"message"`
	Title   string  `json:"title"`
}
