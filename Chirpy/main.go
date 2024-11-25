package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"Chirpy/internal/auth"
	"Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	dbQueries      *database.Queries
	platform       string
	token          string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type Token struct {
	Token string `json:"token"`
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type e struct {
		Err string `json:"error"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	id, err := auth.ValidateJWT(token, []byte(cfg.token))
	if err != nil {
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	params.UserID = id
	err = decoder.Decode(&params)
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
		chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: fixedChirp, UserID: params.UserID})
		if err != nil {
		}
		r := database.Res{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: chirp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		dat, err := json.Marshal(r)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
		w.Write(dat)
		return
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: params.UserID})
	if err != nil {
	}
	r2 := database.Res{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: chirp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	dat, err := json.Marshal(r2)
	if err != nil {
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.ListChirps(r.Context())
	if err != nil {
	}
	res := database.MapSqlChirpsToJsonChirps(chirps)
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(200)
	c, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(c)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := r.PathValue("chirpID")
		chirp, err := cfg.dbQueries.ListChirp(r.Context(), uuid.MustParse(id))
		if err != nil {
			w.WriteHeader(404)
			return
		}
		res := database.MapSqlChirpToJsonChirp(chirp)
		w.WriteHeader(200)
		c, err := json.Marshal(res)
		if err != nil {
		}
		w.Write(c)
	}

	if r.Method == http.MethodDelete {
		cfg.deleteChirpById(w, r)
	}
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
	}
	hashedPass, err := auth.HashedPassword(params.Password)
	if err != nil {
	}
	userCred := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), userCred)
	if err != nil {
	}
	newUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	userJSON, err := json.Marshal(newUser)
	if err != nil {
	}
	w.WriteHeader(201)
	w.Write(userJSON)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
	}
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := request{}
	err = decoder.Decode(&params)
	if err != nil {
	}
	hashedPass, err := auth.HashedPassword(params.Password)
	if err != nil {
	}
	userID, err := auth.ValidateJWT(authToken, []byte(cfg.token))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	user, err := cfg.dbQueries.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{Email: params.Email, HashedPassword: hashedPass, ID: userID})
	if err != nil {
	}
	newUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	userJSON, err := json.Marshal(newUser)
	if err != nil {
	}
	w.WriteHeader(200)
	w.Write(userJSON)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
	}
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	hashErr := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if hashErr != nil {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	expiresInSeconds := 3600
	t, err := auth.MakeJWT(user.ID, []byte(cfg.token), time.Duration(expiresInSeconds)*time.Second)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
	}
	// insert refresh token into db
	rt, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refreshToken, UserID: user.ID})
	if err != nil {
	}
	newUser := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        t,
		RefreshToken: rt.Token,
	}
	userJSON, err := json.Marshal(newUser)
	if err != nil {
	}
	w.WriteHeader(200)
	w.Write(userJSON)
}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
	}
	rf, err := cfg.dbQueries.GetRefreshToken(r.Context(), authToken)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if rf.RevokedAt.Valid {
		w.WriteHeader(401)
		return
	}
	userID, err := cfg.dbQueries.GetUserByRefreshToken(r.Context(), rf.Token)
	if err != nil {
	}
	if userID == uuid.Nil {
		w.WriteHeader(401)
		return
	}
	expiresInSeconds := 3600
	t, err := auth.MakeJWT(userID, []byte(cfg.token), time.Duration(expiresInSeconds)*time.Second)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	newToken := Token{
		Token: t,
	}
	ret, err := json.Marshal(newToken)
	if err != nil {
	}
	w.WriteHeader(200)
	w.Write(ret)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
	}
	err = cfg.dbQueries.UpdateRefreshToken(r.Context(), authToken)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) resetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	cfg.dbQueries.ResetUsers(r.Context())
}

func (cfg *apiConfig) deleteChirpById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	chirp, err := cfg.dbQueries.ListChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		w.WriteHeader(404)
		return
	}
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
	}
	userID, err := auth.ValidateJWT(authToken, []byte(cfg.token))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if chirp.UserID != userID {
		w.WriteHeader(403)
		return
	}
	err = cfg.dbQueries.DeleteChirpById(r.Context(), uuid.MustParse(id))
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handleWebhooks(w http.ResponseWriter, r *http.Request) {

}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal()
	}
	dbQueries := database.New(db)
	mux := http.NewServeMux()
	apiConfig := apiConfig{
		fileserverHits: 0,
		dbQueries:      dbQueries,
		platform:       platform,
		token:          tokenSecret,
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
	mux.HandleFunc("GET /api/chirps", apiConfig.getChirps)
	mux.HandleFunc("/api/chirps/{chirpID}", apiConfig.getChirpById)
	mux.HandleFunc("POST /api/users", apiConfig.createUser)
	mux.HandleFunc("PUT /api/users", apiConfig.updateUser)
	mux.HandleFunc("POST /admin/reset", apiConfig.resetUsers)
	mux.HandleFunc("POST /api/chirps", apiConfig.createChirp)
	mux.HandleFunc("POST /api/login", apiConfig.loginUser)
	mux.HandleFunc("POST /api/refresh", apiConfig.refreshToken)
	mux.HandleFunc("POST /api/revoke", apiConfig.revokeToken)
	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handleWebhooks)
	fmt.Println("Server running...")
	server.ListenAndServe()
}
