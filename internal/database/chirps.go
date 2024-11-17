package database

import (
	"sort"

	"github.com/MazzMS/chirpy-rrss/internal/models"
)

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, authorId int) (models.Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return models.Chirp{}, err
	}
	chirpId := dbStructure.LastChirpId + 1
	dbStructure.LastChirpId++
	chirp := models.Chirp{
		Id:   chirpId,
		Body: body,
		AuthorId: authorId,
	}
	dbStructure.Chirps[chirpId] = chirp
	err = db.writeDB(dbStructure)
	if err != nil {
		return models.Chirp{}, err
	}
	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps(authorId int, sortAsc bool) ([]models.Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := []models.Chirp{}
	for _, chirp := range dbStructure.Chirps {
		if authorId == 0 {
			chirps = append(chirps, chirp)
		} else {
			if chirp.AuthorId == authorId {
				chirps = append(chirps, chirp)
			}
		}
	}
	if sortAsc {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	} else {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id > chirps[j].Id })
	}
	return chirps, nil
}

func (db *DB) DeleteChirp(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(dbStructure.Chirps, id)

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}
