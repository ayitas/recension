package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ayitas/recension/api/internal/store"
)

type entityRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type memberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req entityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(req.Name)
	slug := normalizeSlug(req.Slug)
	if slug == "" {
		slug = normalizeSlug(name)
	}
	if name == "" {
		name = slug
	}
	if !validSlug(slug) {
		writeErr(w, http.StatusBadRequest, "slug must be 2-64 chars of lowercase letters, numbers, and hyphens")
		return
	}
	if _, exists := s.store.TeamBySlug(slug); exists {
		writeErr(w, http.StatusConflict, "team already exists")
		return
	}
	team := s.store.EnsureTeam(slug, name)
	if err := s.store.AddMember(team.ID, user.ID, store.RoleOwner); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not assign ownership")
		return
	}
	writeJSON(w, http.StatusCreated, team)
}

func (s *Server) handleCreateSuite(w http.ResponseWriter, r *http.Request) {
	teamSlug := r.PathValue("team")
	var req entityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(req.Name)
	slug := normalizeSlug(req.Slug)
	if slug == "" {
		slug = normalizeSlug(name)
	}
	if name == "" {
		name = slug
	}
	if !validSlug(slug) {
		writeErr(w, http.StatusBadRequest, "slug must be 2-64 chars of lowercase letters, numbers, and hyphens")
		return
	}
	if _, ok := s.store.TeamBySlug(teamSlug); !ok {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	for _, suite := range s.store.ListSuites(teamSlug) {
		if suite.Slug == slug {
			writeErr(w, http.StatusConflict, "suite already exists")
			return
		}
	}
	suite, err := s.store.EnsureSuite(teamSlug, slug, name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, suite)
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.store.ListMembers(r.PathValue("team"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (s *Server) handleAddMember(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	teamSlug := r.PathValue("team")
	team, ok := s.store.TeamBySlug(teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	actorRole, _ := s.store.Membership(user.ID, teamSlug)

	var req memberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if role == "" {
		role = store.RoleMember
	}
	if !store.ValidRole(role) {
		writeErr(w, http.StatusBadRequest, "role must be owner, admin, member, or viewer")
		return
	}
	if role == store.RoleOwner && actorRole != store.RoleOwner {
		writeErr(w, http.StatusForbidden, "only owners can grant owner role")
		return
	}
	target, ok := s.store.UserByEmail(email)
	if !ok {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	if existing, ok := s.store.Membership(target.ID, teamSlug); ok {
		if existing == store.RoleOwner && actorRole != store.RoleOwner {
			writeErr(w, http.StatusForbidden, "only owners can modify owners")
			return
		}
	}
	if err := s.store.AddMember(team.ID, target.ID, role); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, store.TeamMember{
		UserID: target.ID,
		Email:  target.Email,
		Role:   role,
	})
}

func (s *Server) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	teamSlug := r.PathValue("team")
	targetID := r.PathValue("userId")
	team, ok := s.store.TeamBySlug(teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	actorRole, _ := s.store.Membership(user.ID, teamSlug)

	var req memberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if !store.ValidRole(role) {
		writeErr(w, http.StatusBadRequest, "role must be owner, admin, member, or viewer")
		return
	}
	targetRole, ok := s.store.Membership(targetID, teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "member not found")
		return
	}
	if targetRole == store.RoleOwner && actorRole != store.RoleOwner {
		writeErr(w, http.StatusForbidden, "only owners can modify owners")
		return
	}
	if role == store.RoleOwner && actorRole != store.RoleOwner {
		writeErr(w, http.StatusForbidden, "only owners can grant owner role")
		return
	}
	if err := s.store.SetMemberRole(team.ID, targetID, role); err != nil {
		if errors.Is(err, store.ErrLastOwner) {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	target, _ := s.store.UserByID(targetID)
	email := ""
	if target != nil {
		email = target.Email
	}
	writeJSON(w, http.StatusOK, store.TeamMember{UserID: targetID, Email: email, Role: role})
}

func (s *Server) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	teamSlug := r.PathValue("team")
	targetID := r.PathValue("userId")
	team, ok := s.store.TeamBySlug(teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	actorRole, _ := s.store.Membership(user.ID, teamSlug)
	targetRole, ok := s.store.Membership(targetID, teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "member not found")
		return
	}
	if targetRole == store.RoleOwner && actorRole != store.RoleOwner {
		writeErr(w, http.StatusForbidden, "only owners can remove owners")
		return
	}
	if err := s.store.RemoveMember(team.ID, targetID); err != nil {
		if errors.Is(err, store.ErrLastOwner) {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
