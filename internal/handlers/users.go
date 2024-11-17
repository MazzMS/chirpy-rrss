package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
	"github.com/MazzMS/chirpy-rrss/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func NewUser(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING newUser")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}
	handleError := func(err error, msg string) {
		if msg == "" {
			msg = "Something went wrong"
		}
		log.Printf("Error: %s", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	// initialize vars
	param := parameter{}
	res := response{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)

	if err != nil {
		handleError(err, "")
		return
	}

	// GenerateFromPassword() requires that the password be no longer than 72 bytes
	if len([]byte(param.Password)) > 72 {
		handleError(err, "Password cannot be longer than 72 bytes")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if param.Email == "" {
		handleError(err, "Email cannot be empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "")
		return
	}

	// check if email already registered
	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "")
		return
	}
	if _, ok := users[param.Email]; ok {
		handleError(err, "email already in use")
		return
	}

	email := param.Email
	password, err := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
	if err != nil {
		handleError(err, "")
		return
	}

	user, err := db.CreateUser(email, password)
	if err != nil {
		handleError(err, "")
		return
	}
	res.Email = user.Email
	res.Id = user.Id
	res.IsChirpyRed = user.IsChirpyRed

	data, err := json.Marshal(res)
	if err != nil {
		handleError(err, "")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
	return
}

func GetUser(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING getUser")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type parameter struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}
	handleError := func(err error, msg string) {
		if msg == "" {
			msg = "Something went wrong"
		}
		if config.Debug {
			log.Printf("Error: %s", err)
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	// initialize vars
	param := parameter{}
	res := response{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)

	if err != nil {
		handleError(err, "")
		return
	}

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "")
		return
	}

	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "")
		return
	}

	user, ok := users[param.Email]
	if !ok {
		handleError(err, "")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	res.Email = user.Email
	res.Id = user.Id
	res.IsChirpyRed = user.IsChirpyRed

	data, err := json.Marshal(res)
	if err != nil {
		handleError(err, "")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func Login(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING login")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type parameter struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		Id           int    `json:"id"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
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

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "", 0)
		return
	}

	user, ok := users[param.Email]
	// if user do not exist
	if !ok {
		handleError(err, "user not found", http.StatusNotFound)
		return
	}
	// if hashes do not match
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(param.Password))
	if err != nil {
		handleError(err, "", http.StatusUnauthorized)
		return
	}

	// jwt
	if config.Debug {
		log.Printf("User id: %d. Stringified %s\n", user.Id, strconv.Itoa(user.Id))
	}
	const defaultExpirationTime = 60 * 60
	if param.ExpiresInSeconds <= 0 || param.ExpiresInSeconds > defaultExpirationTime {
		param.ExpiresInSeconds = defaultExpirationTime
	}
	currentTime := time.Now().UTC()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(param.ExpiresInSeconds) * time.Second)),
			Subject:   strconv.Itoa(user.Id),
		})
	signed, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		handleError(err, "", 0)
		return
	}
	// refresh token
	refreshToken := make([]byte, 32)
	_, err = rand.Read(refreshToken)
	if err != nil {
		handleError(err, "", 0)
		return
	}
	encodedRefreshToken := hex.EncodeToString(refreshToken)

	expirationDuration := 60 * 24 * time.Hour // 60 days
	expiresAt := time.Now().UTC().Add(expirationDuration)

	recordedRefreshToken, err := db.CreateRefreshToken(encodedRefreshToken, param.Email, expiresAt)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	res.Email = user.Email
	res.Id = user.Id
	res.Token = signed
	res.RefreshToken = recordedRefreshToken.Token
	res.IsChirpyRed = user.IsChirpyRed

	data, err := json.Marshal(res)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func UpdateUser(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING updateUser")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Email       string `json:"email"`
		Id          int    `json:"id"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
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

	// check jwt
	possibleToken := r.Header.Get("Authorization")
	if config.Debug {
		log.Printf("String token: %s\n", possibleToken)
	}
	if possibleToken == "" {
		handleError(fmt.Errorf("Header did not contain Authorization: %v", r.Header), "", http.StatusUnauthorized)
		return
	}
	possibleToken = possibleToken[len("Bearer "):]
	token, err := jwt.ParseWithClaims(
		possibleToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(config.JwtSecret), nil },
	)
	if err != nil {
		handleError(fmt.Errorf("bad token: %v", err), "", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		handleError(fmt.Errorf("unknown claim types, cannot proceed"), "", http.StatusUnauthorized)
		return
	}
	if config.Debug {
		log.Printf("Claims: %s\n", claims)
	}

	// get id
	if config.Debug {
		log.Printf("Subject: %s\n", claims.Subject)
	}
	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	res := response{}
	param := params{}

	// decode input
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&param)

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// check if email is null
	if param.Email == "" {
		handleError(fmt.Errorf("email is null\n"), "", 0)
		return
	}
	// check if email already in use
	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "", 0)
		return
	}
	if _, ok := users[param.Email]; ok {
		handleError(fmt.Errorf("email already in use\n"), "", 0)
		return
	}

	// check if pass is null
	if param.Password == "" {
		handleError(fmt.Errorf("pass is null\n"), "", 0)
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	// update user
	user, err := db.UpdateUser(id, param.Email, hashed, false)
	if err != nil {
		handleError(err, "", http.StatusInternalServerError)
		return
	}

	res.Email = user.Email
	res.Id = user.Id
	res.IsChirpyRed = user.IsChirpyRed

	data, err := json.Marshal(res)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
