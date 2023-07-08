package cache

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"time"
)

type UserVk struct {
	Token *oauth2.Token `json:"token"`
	Info  *UserInfo     `json:"info"`
}

func (u UserVk) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u UserVk) Valid() bool {
	timeNow := time.Now()
	exp := u.Token.Expiry
	after := exp.After(timeNow)
	return after
}

type UserInfo struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	SecondName string `json:"secondName"`
	Photo      string `json:"photo_400_orig"`
}

func (u UserInfo) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}
