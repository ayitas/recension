package httpapi

import (
	"net/http"

	"github.com/ayitas/recension/api/internal/store"
)

// authorizeTeamRole checks that user has at least minRole on teamSlug.
// On failure writes a 404 (to avoid leaking team existence) and returns false.
func (s *Server) authorizeTeamRole(w http.ResponseWriter, user *store.User, teamSlug, minRole string) bool {
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	role, ok := s.store.Membership(user.ID, teamSlug)
	if !ok || !store.RoleAtLeast(role, minRole) {
		writeErr(w, http.StatusNotFound, "team not found")
		return false
	}
	return true
}

func (s *Server) requireTeamRole(minRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := currentUser(r)
		if !s.authorizeTeamRole(w, user, r.PathValue("team"), minRole) {
			return
		}
		next(w, r)
	}
}

func (s *Server) authorizeAnyTeamRole(w http.ResponseWriter, user *store.User, minRole string) bool {
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	for _, team := range s.store.ListTeamsForUser(user.ID) {
		role, ok := s.store.Membership(user.ID, team.Slug)
		if ok && store.RoleAtLeast(role, minRole) {
			return true
		}
	}
	writeErr(w, http.StatusForbidden, "team membership required")
	return false
}
