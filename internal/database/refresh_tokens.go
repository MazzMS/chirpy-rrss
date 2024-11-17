package database

import (
	"fmt"
	"time"

	"github.com/MazzMS/chirpy-rrss/internal/models"
)

func (db *DB) CreateRefreshToken(token, userEmail string, expiresAt time.Time) (models.RefreshToken, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return models.RefreshToken{}, err
	}
	refreshToken := models.RefreshToken{
		Token:     token,
		UserEmail: userEmail,
		ExpiresAt: expiresAt,
	}
	dbStructure.RefreshTokens[token] = refreshToken
	err = db.writeDB(dbStructure)
	if err != nil {
		return models.RefreshToken{}, err
	}
	return refreshToken, nil
}

// GetRefreshTokens returns all tokens in the database
func (db *DB) GetRefreshTokens() (map[string]models.RefreshToken, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	refreshTokens := make(map[string]models.RefreshToken)
	for _, token := range dbStructure.RefreshTokens {
		refreshTokens[token.Token] = token
	}
	return refreshTokens, nil
}

func (db *DB) DeleteRefreshToken(token string) error {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	refreshTokens, err := db.GetRefreshTokens()
	if err != nil {
		return err
	}
	_, ok := refreshTokens[token]
	if !ok {
		return fmt.Errorf("token not found")
	}
	delete(refreshTokens, token)
	dbStructure.RefreshTokens = refreshTokens
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}
