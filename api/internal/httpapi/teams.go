package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

type entityRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizeSlug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func validSlug(s string) bool {
	return len(s) >= 2 && len(s) <= 64 && slugRe.MatchString(s)
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
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
	for _, t := range s.store.ListTeams() {
		if t.Slug == slug {
			writeErr(w, http.StatusConflict, "team already exists")
			return
		}
	}
	team := s.store.EnsureTeam(slug, name)
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
	foundTeam := false
	for _, t := range s.store.ListTeams() {
		if t.Slug == teamSlug {
			foundTeam = true
			break
		}
	}
	if !foundTeam {
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
