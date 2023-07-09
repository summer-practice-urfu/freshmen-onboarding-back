package models

type UserPostRating struct {
	UserId int    `json:"userId"`
	PostId string `json:"postId"`
	Oper   rune   `json:"oper"`
}
