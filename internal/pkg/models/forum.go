package models

//easyjson:json
type Forum struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}

//easyjson:json
type Thread struct {
	ID      int     `json:"id"`
	Forum   string  `json:"forum"`
	Title   string  `json:"title"`
	Author  string  `json:"author"`
	Message string  `json:"message"`
	Slug    *string `json:"slug,omitempty"`
	Created *string `json:"created,omitempty"`
}
