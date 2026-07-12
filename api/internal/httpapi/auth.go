package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ayitas/recension/api/internal/auth"
	"github.com/ayitas/recension/api/internal/store"
)

type contextKey string

const userContextKey contextKey = "user"

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.BearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := s.tokens.Parse(tok)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		user, ok := s.store.UserByID(claims.UserID)
		if !ok {
			writeErr(w, http.StatusUnauthorized, "user not found")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	}
}

func currentUser(r *http.Request) *store.User {
	u, _ := r.Context().Value(userContextKey).(*store.User)
	return u
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	if !s.config.AllowSignup {
		writeErr(w, http.StatusForbidden, "signup disabled")
		return
	}
	if !s.authLimiter.allow("signup:" + clientIP(r)) {
		writeErr(w, http.StatusTooManyRequests, "too many requests")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") || len(req.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "email required and password must be at least 8 characters")
		return
	}
	if _, exists := s.store.UserByEmail(email); exists {
		writeErr(w, http.StatusConflict, "email already registered")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	apiKey, err := auth.NewAPIKey()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create api key")
		return
	}
	user, err := s.store.CreateUser(email, hash, apiKey)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not create user")
		return
	}
	s.writeSession(w, user)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.authLimiter.allow("login:" + clientIP(r)) {
		writeErr(w, http.StatusTooManyRequests, "too many requests")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	user, ok := s.store.UserByEmail(email)
	if !ok || user.PasswordHash == "" || !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	s.writeSession(w, user)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":     user.ID,
		"email":  user.Email,
		"apiKey": user.APIKey,
	})
}

func (s *Server) handleRotateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	newKey, err := auth.NewAPIKey()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create api key")
		return
	}
	updated, err := s.store.RotateAPIKey(user.ID, newKey)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":     updated.ID,
		"email":  updated.Email,
		"apiKey": updated.APIKey,
	})
}

func (s *Server) writeSession(w http.ResponseWriter, user *store.User) {
	token, exp, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not issue session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"expiresAt": exp.UTC(),
		"user": map[string]any{
			"id":     user.ID,
			"email":  user.Email,
			"apiKey": user.APIKey,
		},
	})
}
