package database

import (
	"fmt"

	"github.com/MazzMS/chirpy-rrss/internal/models"
)

func (db *DB) CreateUser(email string, password []byte) (models.User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return models.User{}, err
	}
	userId := dbStructure.LastUserId + 1
	dbStructure.LastUserId++
	user := models.User{
		Id:       userId,
		Email:    email,
		Password: password,
		IsChirpyRed: false,
	}
	dbStructure.Users[userId] = user
	err = db.writeDB(dbStructure)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (db *DB) UpdateUser(userId int, email string, password []byte, isChirpyRed bool) (models.User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return models.User{}, err
	}

	// get old user
	oldUser, ok := dbStructure.Users[userId]
	if !ok {
		return models.User{}, fmt.Errorf("user id not found")
	}

	user := models.User{
		Id:       userId,
		Email:    email,
		Password: password,
		IsChirpyRed: isChirpyRed,
	}

	dbStructure.Users[userId] = user

	// change email in refresh tokens
	for token, refreshToken := range dbStructure.RefreshTokens {
		if refreshToken.UserEmail == oldUser.Email {
			refreshToken.UserEmail = user.Email
			dbStructure.RefreshTokens[token] = refreshToken
		}
	}

	err = db.writeDB(dbStructure)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// GetUsers returns all users in the database
func (db *DB) GetUsers() (map[string]models.User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	users := make(map[string]models.User)
	for _, user := range dbStructure.Users {
		users[user.Email] = user
	}
	return users, nil
}
