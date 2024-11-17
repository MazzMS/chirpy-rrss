package database

import (
	"os"
	"sync"

	models "github.com/MazzMS/chirpy-rrss/internal/models"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]models.Chirp           `json:"chirps"`
	LastChirpId   int                            `json:"last_chirp_id"`
	Users         map[int]models.User            `json:"users"`
	LastUserId    int                            `json:"last_user_id"`
	RefreshTokens map[string]models.RefreshToken `json:"refresh_tokens"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(filename string) (*DB, error) {
	database := &DB{
		filename,
		&sync.RWMutex{},
	}
	err := database.ensureDB()
	if err != nil {
		return nil, err
	}

	return database, nil
}

// DeleteDB deletes the DB file
func DeleteDB(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if os.IsNotExist(err) {
		err = os.WriteFile(
			db.path,
			[]byte("{\"chirps\": {}, \"last_chirp_id\": 0, \"users\": {}, \"last_user_id\": 0, \"refresh_tokens\": {}}"),
			0644,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

