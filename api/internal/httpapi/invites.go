package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ayitas/recension/api/internal/store"
)

type inviteRequest struct {
	Role string `json:"role"`
}

func (s *Server) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	teamSlug := r.PathValue("team")
	team, ok := s.store.TeamBySlug(teamSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "team not found")
		return
	}
	actorRole, _ := s.store.Membership(user.ID, teamSlug)

	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
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
	inv, err := s.store.CreateInvite(team.ID, user.ID, role, 7*24*time.Hour)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"token":     inv.Token,
		"role":      inv.Role,
		"expiresAt": inv.ExpiresAt.UTC(),
		"path":      "/invite/" + inv.Token,
	})
}

func (s *Server) handleGetInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	inv, team, err := s.store.InviteByToken(token)
	if err != nil {
		writeErr(w, http.StatusNotFound, "invite not found")
		return
	}
	expired := inv.AcceptedAt != nil || time.Now().UTC().After(inv.ExpiresAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"team":      team,
		"role":      inv.Role,
		"expiresAt": inv.ExpiresAt.UTC(),
		"expired":   expired,
	})
}

func (s *Server) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	token := r.PathValue("token")
	member, err := s.store.AcceptInvite(token, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrInviteExpired) {
			writeErr(w, http.StatusGone, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	inv, team, _ := s.store.InviteByToken(token)
	_ = inv
	writeJSON(w, http.StatusOK, map[string]any{
		"member": member,
		"team":   team,
	})
}
