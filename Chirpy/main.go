package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) increaseHits() {
	cfg.fileserverHits = cfg.fileserverHits + 1
}

func (cfg *apiConfig) checkHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`, cfg.fileserverHits)))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.increaseHits()
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type fixed struct {
		Clean string `json:"cleaned_body"`
	}

	type e struct {
		Err string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		er := e{
			Err: "Something went wrong",
		}
		dat, err := json.Marshal(er)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {
		er := e{
			Err: "Chirp is too long",
		}
		dat, err := json.Marshal(er)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	words := strings.Split(params.Body, " ")
	isFixed := false
	var newWords []string
	for _, word := range words {
		w := strings.ToLower(word)
		if w == "kerfuffle" || w == "sharbert" || w == "fornax" {
			word = "****"
			isFixed = true
			newWords = append(newWords, word)
		} else {
			newWords = append(newWords, word)
		}
	}

	if isFixed {
		fixedChirp := strings.Join(newWords, " ")
		r := fixed{
			Clean: fixedChirp,
		}
		dat, err := json.Marshal(r)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
		return
	}

	r2 := fixed{
		Clean: string(params.Body),
	}
	dat, err := json.Marshal(r2)
	if err != nil {
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(db)
	chirps, err := db.GetChirps()
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(200)
	c, err := json.Marshal(chirps)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(c)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal()
	}
	dbQueries := database.New(db)
	log.Println(dbQueries)
	mux := http.NewServeMux()
	apiConfig := apiConfig{
		fileserverHits: 0,
	}
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiConfig.checkHits)
	mux.HandleFunc("/api/reset", apiConfig.resetHits)
	mux.HandleFunc("POST /api/chirps", apiConfig.createChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.getChirps)
	fmt.Println("Server running...")
	server.ListenAndServe()
}
