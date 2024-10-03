package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dis012/ChirpyWebServer/internal"
	"github.com/dis012/ChirpyWebServer/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`         // Change to lowercase 'id'
	CreatedAt time.Time `json:"created_at"` // Change to 'created_at'
	UpdatedAt time.Time `json:"updated_at"` // Change to 'updated_at'
	UserID    uuid.UUID `json:"user_id"`    // Change to 'user_id'
	Body      string    `json:"body"`       // Change to 'body'
}

type User struct {
	Email    string `json:"email"`    // Change to 'email'
	Password string `json:"password"` // Change to 'hashed_password'
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Handler that writes the hit counter to the response as a text/plain response
func (a *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, a.fileServerHits.Load())))
}

func (a *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if a.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset metrics is only allowed in dev mode"))
		return
	}

	err := a.dbQueries.DeleteAllTokens(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.dbQueries.DeleteAllChirps(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All has been deleted"))
}

func (a *apiConfig) createNewUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashed_password, err := internal.HashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUser, err := a.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "created_at": "%s", "updated_at": "%s", "email": "%s"}`, newUser.ID, newUser.CreatedAt, newUser.UpdatedAt, newUser.Email)))
}

func (a *apiConfig) createNewChirpHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err := internal.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := internal.ValidateJWT(tokenString, a.secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	chirpParam := jsonHandlerForChirp(w, r)

	databaseChirpParam := database.CreateChirpParams{
		Body:   chirpParam.Body,
		UserID: userID,
	}

	chirp, err := a.dbQueries.CreateChirp(r.Context(), databaseChirpParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "created_at": "%s", "updated_at": "%s", "user_id": "%s", "body": "%s"}`, chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.UserID, chirp.Body)))
}

func (a *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := a.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(chirps) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var chirpsSet []Chirp

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	for _, chirp := range chirps {
		chirpsSet = append(chirpsSet, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
		})
	}
	// Write the array of chirps to the response
	json.NewEncoder(w).Encode(chirpsSet)
}

func (a *apiConfig) getChirpByIdHandler(w http.ResponseWriter, r *http.Request) {
	ChirpParam, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	chirp, err := a.dbQueries.GetChirpById(r.Context(), ChirpParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// if chirp is not found, return 404
	if chirp.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "created_at": "%s", "updated_at": "%s", "user_id": "%s", "body": "%s"}`, chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.UserID, chirp.Body)))
}

func (a *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	var userData User
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := a.dbQueries.GetUserByEmail(r.Context(), userData.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.ID == uuid.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	checkIfPasswordCorrect := internal.CheckPassword(userData.Password, user.HashedPassword)

	if !checkIfPasswordCorrect {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Access token expires in 1h
	token, err := internal.MakeJWT(user.ID, a.secret, time.Hour*1)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := internal.MakeRefreshToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
	}

	refreshToken, err := a.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"id": "%s", "created_at": "%s", "updated_at": "%s", "email": "%s", "token": "%s", "refresh_token": "%s"}`, user.ID, user.CreatedAt, user.UpdatedAt, user.Email, token, refreshToken.Token)))
}

func (a *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	requestToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	refreshToken, err := a.dbQueries.GetRefreshToken(r.Context(), requestToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		http.Error(w, "token expired", http.StatusUnauthorized)
		return
	}

	if refreshToken.RevokedAt.Valid {
		// The token is revoked, you can handle accordingly
		http.Error(w, "Refresh token is revoked", http.StatusUnauthorized)
		return
	}

	// Access token expires in 1h
	token, err := internal.MakeJWT(refreshToken.UserID, a.secret, time.Hour*1)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
}

func (a *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	requstToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, err := a.dbQueries.GetRefreshToken(r.Context(), requstToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = a.dbQueries.RevokeRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
