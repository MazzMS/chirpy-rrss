package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/MazzMS/chirpy-rrss/internal/database"
	"github.com/MazzMS/chirpy-rrss/internal/handlers"
	cfg "github.com/MazzMS/chirpy-rrss/internal/config"
	dotenv "github.com/joho/godotenv"
)

func main() {
	const port = "8080"
	const filepathRoot = "."
	var config cfg.ApiConfig
	wrapper := func(handler func(http.ResponseWriter, *http.Request, *cfg.ApiConfig), config *cfg.ApiConfig) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, config)
		}
	}

	dotenv.Load()
	config.JwtSecret = os.Getenv("JWT_SECRET")
	config.PolkaApiKey = os.Getenv("POLKA_API_KEY")

	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	config.Debug = *debug

	if config.Debug {
		log.Println("Using debug mode")
		database.DeleteDB("database.json")
	}

	mux := http.NewServeMux()
	mux.Handle(
		"GET /app/*",
		config.MiddlewereMetricsInt(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))),
	)
	mux.HandleFunc("GET /api/healthz", handlers.Healthz)
	// metrics
	mux.HandleFunc(
		"GET /admin/metrics",
		wrapper(handlers.Metrics, &config),
	)
	mux.HandleFunc("GET /api/reset", wrapper(handlers.Reset, &config))
	// chirps
	mux.HandleFunc("POST /api/chirps", wrapper(handlers.NewChirp, &config))
	mux.HandleFunc("GET /api/chirps", wrapper(handlers.GetChirps, &config))
	mux.HandleFunc("GET /api/chirps/{chirpId}", wrapper(handlers.GetChirp, &config))
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", wrapper(handlers.DeleteChirp, &config))
	// users
	mux.HandleFunc("POST /api/users", wrapper(handlers.NewUser, &config))
	mux.HandleFunc("GET /api/users/{userId}", wrapper(handlers.GetUser, &config))
	mux.HandleFunc("POST /api/login", wrapper(handlers.Login, &config))
	mux.HandleFunc("PUT /api/users", wrapper(handlers.UpdateUser, &config))
	// token
	mux.HandleFunc("POST /api/refresh", wrapper(handlers.NewToken, &config))
	mux.HandleFunc("POST /api/revoke", wrapper(handlers.DeleteToken, &config))
	// chirpy red
	mux.HandleFunc("POST /api/polka/webhooks", wrapper(handlers.PolkaWebhook, &config))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
