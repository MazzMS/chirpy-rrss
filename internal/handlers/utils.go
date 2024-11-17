package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
)

func getIdJwt(r *http.Request, config *cfg.ApiConfig) (int, error) {
	possibleToken := r.Header.Get("Authorization")
	if config.Debug {
		log.Printf("String token: %s\n", possibleToken)
	}
	if possibleToken == "" {
		return 0, fmt.Errorf("Header did not contain Authorization: %v", r.Header)
	}
	possibleToken = possibleToken[len("Bearer "):]
	token, err := jwt.ParseWithClaims(
		possibleToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(config.JwtSecret), nil },
	)
	if err != nil {
		return 0, fmt.Errorf("bad token: %v", err)
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, fmt.Errorf("unknown claim types, cannot proceed")
	}
	if config.Debug {
		log.Printf("Claims: %s\n", claims)
	}

	// get id
	if config.Debug {
		log.Printf("Subject: %s\n", claims.Subject)
	}
	authorId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, err
	}

	return authorId, nil
}
