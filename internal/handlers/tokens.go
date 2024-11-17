package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
	"github.com/MazzMS/chirpy-rrss/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

func NewToken(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING newToken")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type response struct {
		Token string `json:"token"`
	}
	handleError := func(err error, msg string, code int) {
		if msg == "" {
			msg = "Something went wrong"
		}
		if code == 0 {
			code = http.StatusUnauthorized
		}
		if config.Debug {
			log.Printf("Error: %s", err)
			log.Print("FAIL")
		}
		http.Error(w, msg, code)
	}

	// initialize vars
	res := response{}

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	refreshTokens, err := db.GetRefreshTokens()
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// extract token from header
	possibleToken := r.Header.Get("Authorization")
	if config.Debug {
		log.Printf("String token: %s", possibleToken)
	}
	if possibleToken == "" {
		handleError(fmt.Errorf("Header did not contain Authorization: %v", r.Header), "", 0)
		return
	}
	possibleToken = possibleToken[len("Bearer "):]

	// check token
	refreshToken, ok := refreshTokens[possibleToken]
	// if token do not exist
	if !ok {
		handleError(err, "user not found", 0)
		return
	}
	// if token expired
	if time.Now().UTC().Compare(refreshToken.ExpiresAt) < -1 {
		// delete from DB
		err := db.DeleteRefreshToken(refreshToken.Token)
		if err != nil {
			handleError(err, "", 0)
			return
		}
		handleError(fmt.Errorf("token expired"), "", 0)
		return
	}

	// get user
	users, err := db.GetUsers()
	if err != nil {
		handleError(err, "", 0)
		return
	}
	user, ok := users[refreshToken.UserEmail]
	if config.Debug {
		log.Printf("Current user email: %s\n", refreshToken.UserEmail)
		log.Println("Users in system:")
		for _, user := range users {
			log.Printf("%d: %s\n", user.Id, user.Email)
		}
	}
	if !ok {
		handleError(fmt.Errorf("user not in db"), "", 0)
		return
	}

	// jwt
	if config.Debug {
		log.Printf("User id: %d. Stringified %s\n", user.Id, strconv.Itoa(user.Id))
	}
	const defaultExpirationTime = 60 * 60
	currentTime := time.Now().UTC()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(time.Duration(defaultExpirationTime) * time.Second)),
			Subject:   strconv.Itoa(user.Id),
		})
	signed, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		handleError(err, "", 0)
		return
	}

	res.Token = signed

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

func DeleteToken(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	if config.Debug {
		log.Println()
		log.Println("STARTING deleteToken")
		defer log.Println("FINISH")
		defer fmt.Println()
	}
	// Types for JSON's input and output
	type response struct {
		Token string `json:"token"`
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

	// db interaction
	db, err := database.NewDB("database.json")
	if err != nil {
		handleError(err, "", 0)
		return
	}

	refreshTokens, err := db.GetRefreshTokens()
	if err != nil {
		handleError(err, "", 0)
		return
	}

	// extract token from header
	possibleToken := r.Header.Get("Authorization")
	if config.Debug {
		log.Printf("String token: %s\n", possibleToken)
	}
	if possibleToken == "" {
		handleError(fmt.Errorf("Header did not contain Authorization: %v", r.Header), "", 0)
		return
	}
	possibleToken = possibleToken[len("Bearer "):]

	// check token
	_, ok := refreshTokens[possibleToken]
	// if token do not exist
	if !ok {
		handleError(err, "user not found", 0)
		return
	}

	err = db.DeleteRefreshToken(possibleToken)
	if err != nil {
		handleError(err, "", 0)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
