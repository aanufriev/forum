package models

import "time"

//easyjson:json
type Forum struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Threads int    `json:"threads"`
	Posts   int    `json:"posts"`
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
	Votes   int     `json:"votes"`
}

//easyjson:json
type Post struct {
	ID       int       `json:"id"`
	Author   string    `json:"author"`
	Message  string    `json:"message"`
	Parent   int       `json:"parent,omitempty"`
	Forum    string    `json:"forum"`
	Slug     string    `json:"slug"`
	Thread   int       `json:"thread"`
	Created  time.Time `json:"created,omitempty"`
	IsEdited bool      `json:"isEdited"`
}

//easyjson:json
type PostWrapper struct {
	Post Post `json:"post"`
}

//easyjson:json
type ServiceInfo struct {
	Forum  int `json:"forum"`
	Post   int `json:"post"`
	Thread int `json:"thread"`
	User   int `json:"user"`
}
