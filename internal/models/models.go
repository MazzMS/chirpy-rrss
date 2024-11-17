package models

import "time"

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    []byte `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

type RefreshToken struct {
	Token     string    `json:"token"`
	UserEmail string    `json:"user_email"`
	ExpiresAt time.Time `json:"expires_at"`
}
