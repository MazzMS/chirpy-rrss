package handlers

import (
	"fmt"
	"net/http"

	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
)

func Metrics(w http.ResponseWriter, r *http.Request, config *cfg.ApiConfig) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf(
		"<html><body><h1>Welcome, Chirpy Adming</h1><p>Chirpy has been visited %d times!<p></body></html>",
		config.FileserverHits,
	)
	w.Write([]byte(hits))
}

func Reset(res http.ResponseWriter, req *http.Request, cfg *cfg.ApiConfig) {
	cfg.FileserverHits = 0
	res.WriteHeader(http.StatusOK)
}
