package config

import "net/http"

type ApiConfig struct {
	JwtSecret      string
	PolkaApiKey    string
	FileserverHits int
	Debug          bool
}

func (c *ApiConfig) MiddlewereMetricsInt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		c.FileserverHits++
		res.Header().Add("Cache-Control", "no-cache")
		next.ServeHTTP(res, req)
	})
}
