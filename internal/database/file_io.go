package database

import (
	"encoding/json"
	"os"
)

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	structure := &DBStructure{}
	content, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	err = json.Unmarshal(content, structure)
	if err != nil {
		return DBStructure{}, err
	}
	return *structure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	content, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
