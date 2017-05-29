package models

type (
	Post struct {
		ID       string
		Url      string
		Title    string
		SelfText string
		Comments []string
	}
)

