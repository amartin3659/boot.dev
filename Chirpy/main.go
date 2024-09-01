package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type fixed struct {
		Clean string `json:"cleaned_body"`
	}

	type response struct {
		Valid bool `json:"valid"`
	}

	type e struct {
		Err string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		// log.Printf("Error decoding parameters: %s", err)
		er := e{
			Err: "Something went wrong",
		}
		dat, err := json.Marshal(er)
		if err != nil {
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

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
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
	mux.HandleFunc("POST /api/validate_chirp", apiConfig.validateChirp)
	fmt.Println("Server running...")
	server.ListenAndServe()
}
