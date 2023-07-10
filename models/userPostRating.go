package models

type UserPostRating struct {
	UserId int64  `json:"userId"`
	PostId string `json:"postId"`
	Oper   rune   `json:"oper"`
}
