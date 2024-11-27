package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPasswordHash(password string, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret []byte, expiresIn time.Duration) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(tokenSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func ValidateJWT(tokenString string, tokenSecret []byte) (uuid.UUID, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tokenSecret, nil
	})
	// Check if token is valid and cast claims
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	id, err := claims.GetSubject()
	log.Println("id", id)
	if err != nil {
		log.Println(err)
	}
	uuidId, err := uuid.Parse(id)
	if err != nil {
		log.Println(err)
	}
	return uuidId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	const bearerPrefix = "Bearer "

	// Get the Authorization header value
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Check if the header starts with "Bearer " and extract the token
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header is not a Bearer token")
	}

	// Return the token after "Bearer " prefix
	return strings.TrimPrefix(authHeader, bearerPrefix), nil
}

func MakeRefreshToken() (string, error) {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := hex.EncodeToString(b)
	return s, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Check if the header starts with "ApiKey "
	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("authorization header is not in the correct format")
	}

	// Extract and return the key after "ApiKey "
	return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix)), nil
}
