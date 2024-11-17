package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
	"github.com/MazzMS/chirpy-rrss/internal/database"
	"github.com/MazzMS/chirpy-rrss/internal/utils"
)

func NewChirp(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	// Types for JSON's input and output
	type parameter struct {
		Body string `json:"body"`
	}
	type response struct {
		Error    string `json:"error"`
		Id       int    `json:"id"`
		Body     string `json:"body"`
		AuthorId int    `json:"author_id"`
	}
	handleError := func(err error, msg string, code int) {
		if msg == "" {
			msg = "Something went wrong"
		}
		if code == 0 {
			code = http.StatusInternalServerError
		}
		if config.Debug {
			log.Printf("Error: %s", err)
		}
		http.Error(w, msg, code)
	}
	// initialize vars
	param := parameter{}
	res := response{}
	// decode input
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)

	if err != nil {
		handleError(err, "", 0)
		return
	}

	bodyLen := len(param.Body)
	if bodyLen > 140 {
		res.Error = "Chirp is too long"
		data, err := json.Marshal(res)
		if err != nil {
			handleError(err, "", 0)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}
	if param.Body == "" {
		res.Error = "Chirp cannot be empty"
		data, err := json.Marshal(res)
		if err != nil {
			handleError(err, "", http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// check jwt
	authorId, err := getIdJwt(r, config)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	body := utils.Clean(param.Body, badWords)

	if config.Debug {
		log.Printf("Creating chirp with author_id: %d and body %q", authorId, body)
	}

	chirp, err := db.CreateChirp(body, authorId)
	if err != nil {
		handleError(err, "", 0)
		return
	}
	res.Body = chirp.Body
	res.Id = chirp.Id
	res.AuthorId = chirp.AuthorId

	data, err := json.Marshal(res)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
	return
}

func GetChirp(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	// Types for JSON's output
	handleError := func(err error) {
		log.Printf("Error: %s", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	// get id
	pathValue := r.PathValue("chirpId")
	if pathValue == "" {
		handleError(fmt.Errorf("cannot match wildcard 'chirpId' in path %v", r.URL.Path))
		return
	}
	id, err := strconv.Atoi(pathValue)
	if err != nil {
		handleError(err)
		return
	}

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err)
		return
	}

	chirps, err := db.GetChirps(0, true)
	if err != nil {
		handleError(err)
		return
	}

	if id > len(chirps) {
		w.WriteHeader(http.StatusNotFound)
		if config.Debug {
			log.Printf("Id %d from path %q. Amount of chirps %d\n", id, r.URL.Path, len(chirps))
		}
		return
	}
	// chirps are sorted, so using the index it's ok
	chirp := chirps[id-1]

	data, err := json.Marshal(chirp)
	if err != nil {
		handleError(err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func GetChirps(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	// Types for JSON's output
	handleError := func(err error, msg string, code int) {
		if code == 0 {
			code = http.StatusNotFound
		}
		if msg == "" {
			msg = "Something went wrong"
		}
		if config.Debug {
			log.Printf("Error: %s", err)
		}
		http.Error(w, msg, code)
	}


	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	// get query parameters
	authorIdString := r.URL.Query().Get("author_id")
	sortParameter := r.URL.Query().Get("sort")

	// author
	authorId := 0
	if authorIdString != "" {
		authorId, err = strconv.Atoi(authorIdString)
		if err != nil {
			handleError(err, "author_id is not a number!", http.StatusBadRequest)
		}
	}
	// sort
	sortAsc := true
	if sortParameter != "" {
		if sortParameter == "desc" {
			sortAsc = false
		} else if sortParameter != "asc" {
			handleError(fmt.Errorf("bad sort parameter: %q", sortParameter), "wrong sort", http.StatusBadRequest)
		}
	}


	// get chirps
	chirps, err := db.GetChirps(authorId, sortAsc)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	for _, chirp := range chirps {
		log.Printf("Chirp: %d, with author_id: %d and body: %q", chirp.Id, chirp.AuthorId, chirp.Body)
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func DeleteChirp(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	// Types for JSON's output
	handleError := func(err error, msg string, code int) {
		if msg == "" {
			msg = "Something went wrong"
		}
		if code == 0 {
			code = http.StatusForbidden
		}
		if config.Debug {
			log.Printf("Error: %s", err)
		}
		http.Error(w, msg, code)
	}

	// get chirp_id to delete
	pathValue := r.PathValue("chirpId")
	if pathValue == "" {
		handleError(fmt.Errorf("cannot match wildcard 'chirpId' in path %v", r.URL.Path), "", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(pathValue)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	// get chirps
	chirps, err := db.GetChirps(0, true)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	chirp := chirps[id - 1]

	// get author id from auth
	authorId, err := getIdJwt(r, config)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	// check if author
	if chirp.AuthorId != authorId {
		handleError( fmt.Errorf("user is not author"), "", 0)
		return
	}

	// delete chirp from db
	err = db.DeleteChirp(chirp.Id)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
