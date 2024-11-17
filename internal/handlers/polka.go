package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
	"github.com/MazzMS/chirpy-rrss/internal/database"
)

func PolkaWebhook(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}
	handleError := func(err error, msg string, code int) {
		if msg == "" {
			msg = "Something went wrong"
		}
		if code == 0 {
			code = http.StatusNotFound
		}
		if config.Debug {
			log.Printf("Error: %s", err)
		}
		http.Error(w, msg, code)
	}

	// check if authorized
	possibleToken := r.Header.Get("Authorization")
	if config.Debug {
		log.Printf("Authorization: %s\n", possibleToken)
	}
	if possibleToken == "" {
		handleError(fmt.Errorf("Header did not contain Authorization: %v", r.Header), "", http.StatusUnauthorized)
		return
	}
	possibleToken = possibleToken[len("ApiKey "):]

	if possibleToken != config.PolkaApiKey {
		handleError(fmt.Errorf("Not a valid ApiKey: %q vs %q", possibleToken, config.PolkaApiKey), "", http.StatusUnauthorized)
		return
	}

	// initialize vars
	param := parameter{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// check possible event
	if param.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// get users
	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// try to find user
	for _, user := range users {
		if user.Id == param.Data.UserId {
			_, err := db.UpdateUser(user.Id, user.Email, user.Password, true)
			if err != nil {
				handleError(err, "", 0)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// if not found return 404
	w.WriteHeader(http.StatusNotFound)
	return
}
