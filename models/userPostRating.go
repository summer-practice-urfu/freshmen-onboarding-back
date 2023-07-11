package models

type UserPostRating struct {
	UserId int64  `json:"userId" db:"userId"`
	PostId string `json:"postId" db:"postId"`
	Oper   rune   `json:"oper" db:"oper"`
}
