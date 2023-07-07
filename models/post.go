package models

type Post struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Rating  int    `json:"rating"`
	Img     string `json:"img"`
}
