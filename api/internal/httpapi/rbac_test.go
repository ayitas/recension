package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ayitas/recension/api/internal/auth"
	"github.com/ayitas/recension/api/internal/config"
	"github.com/ayitas/recension/api/internal/httpapi"
	"github.com/ayitas/recension/api/internal/store"
	"github.com/ayitas/recension/pkg/message"
)

func testServer(t *testing.T) (*httpapi.Server, *store.Memory) {
	t.Helper()
	hash, err := auth.HashPassword("dev-password")
	if err != nil {
		t.Fatal(err)
	}
	mem := store.NewMemory("dev-api-key", hash)
	cfg := config.Config{
		AllowSignup:   true,
		SessionSecret: "test-session-secret-for-rbac",
		SessionTTL:    time.Hour,
		MaxBodyBytes:  1 << 20,
		CORSOrigins:   []string{"*"},
	}
	return httpapi.New(mem, cfg, nil), mem
}

func login(t *testing.T, srv *httpapi.Server, email, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.Token == "" {
		t.Fatalf("login token missing: %s", rec.Body.String())
	}
	return out.Token
}

func signup(t *testing.T, srv *httpapi.Server, email, password string) (token, apiKey string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("signup %s: %d %s", email, rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
		User  struct {
			APIKey string `json:"apiKey"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out.Token, out.User.APIKey
}

func TestRBAC_ListTeamsScoped(t *testing.T) {
	srv, _ := testServer(t)
	outsiderTok, _ := signup(t, srv, "outsider@example.com", "password1")

	req := httptest.NewRequest(http.MethodGet, "/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+outsiderTok)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list teams: %d", rec.Code)
	}
	var teams []store.Team
	_ = json.Unmarshal(rec.Body.Bytes(), &teams)
	if len(teams) != 0 {
		t.Fatalf("outsider should see 0 teams, got %d", len(teams))
	}

	ownerTok := login(t, srv, "dev@recension.local", "dev-password")
	req = httptest.NewRequest(http.MethodGet, "/v1/teams", nil)
	req.Header.Set("Authorization", "Bearer "+ownerTok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &teams)
	if len(teams) != 1 || teams[0].Slug != "acme" {
		t.Fatalf("owner should see acme, got %#v", teams)
	}
}

func TestRBAC_DenyCrossTenantRead(t *testing.T) {
	srv, _ := testServer(t)
	outsiderTok, _ := signup(t, srv, "reader@example.com", "password1")

	req := httptest.NewRequest(http.MethodGet, "/v1/teams/acme/suites", nil)
	req.Header.Set("Authorization", "Bearer "+outsiderTok)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-tenant read, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestRBAC_DenyCrossTenantSubmit(t *testing.T) {
	srv, _ := testServer(t)
	_, outsiderKey := signup(t, srv, "submitter@example.com", "password1")

	gpa := 3.9
	env := message.Envelope{
		Version: message.WireVersion,
		Messages: []message.Message{{
			Metadata: message.Metadata{
				Team: "acme", Suite: "students", Version: "v-rbac", Testcase: "alice",
				BuiltAt: time.Now().UTC(),
			},
			Results: []message.Result{{
				Key: "gpa", Kind: message.KindCheck,
				Value: message.Value{Type: message.TypeDouble, Double: &gpa},
			}},
		}},
	}
	body, _ := json.Marshal(env)
	req := httptest.NewRequest(http.MethodPost, "/v1/client/submit", bytes.NewReader(body))
	req.Header.Set("X-Recension-API-Key", outsiderKey)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-tenant submit, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestRBAC_DenyCrossTenantPromote(t *testing.T) {
	srv, _ := testServer(t)
	gpa := 3.9
	env := message.Envelope{
		Version: message.WireVersion,
		Messages: []message.Message{{
			Metadata: message.Metadata{
				Team: "acme", Suite: "students", Version: "v-prom", Testcase: "alice",
				BuiltAt: time.Now().UTC(),
			},
			Results: []message.Result{{
				Key: "gpa", Kind: message.KindCheck,
				Value: message.Value{Type: message.TypeDouble, Double: &gpa},
			}},
		}},
	}
	body, _ := json.Marshal(env)
	req := httptest.NewRequest(http.MethodPost, "/v1/client/submit", bytes.NewReader(body))
	req.Header.Set("X-Recension-API-Key", "dev-api-key")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("seed submit: %d %s", rec.Code, rec.Body.String())
	}

	outsiderTok, _ := signup(t, srv, "promoter@example.com", "password1")
	req = httptest.NewRequest(http.MethodPost, "/v1/batch/acme/students/v-prom/promote", nil)
	req.Header.Set("Authorization", "Bearer "+outsiderTok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-tenant promote, got %d %s", rec.Code, rec.Body.String())
	}

	ownerTok := login(t, srv, "dev@recension.local", "dev-password")
	req = httptest.NewRequest(http.MethodPost, "/v1/batch/acme/students/v-prom/promote", nil)
	req.Header.Set("Authorization", "Bearer "+ownerTok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("owner promote: %d %s", rec.Code, rec.Body.String())
	}
}

func TestRBAC_CreateTeamMakesOwner(t *testing.T) {
	srv, _ := testServer(t)
	tok, _ := signup(t, srv, "creator@example.com", "password1")

	body, _ := json.Marshal(map[string]string{"name": "Beta", "slug": "beta"})
	req := httptest.NewRequest(http.MethodPost, "/v1/teams", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create team: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/teams/beta/suites", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner read own team: %d", rec.Code)
	}

	// Bootstrap user cannot see beta.
	ownerTok := login(t, srv, "dev@recension.local", "dev-password")
	req = httptest.NewRequest(http.MethodGet, "/v1/teams/beta/suites", nil)
	req.Header.Set("Authorization", "Bearer "+ownerTok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other team's suites, got %d", rec.Code)
	}
}

func TestRBAC_MemberCanSubmitViewerCannot(t *testing.T) {
	srv, mem := testServer(t)
	team, _ := mem.TeamBySlug("acme")

	viewerHash, _ := auth.HashPassword("password1")
	viewerKey, _ := auth.NewAPIKey()
	viewer, err := mem.CreateUser("viewer@example.com", viewerHash, viewerKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.AddMember(team.ID, viewer.ID, store.RoleViewer); err != nil {
		t.Fatal(err)
	}

	gpa := 3.2
	env := message.Envelope{
		Version: message.WireVersion,
		Messages: []message.Message{{
			Metadata: message.Metadata{
				Team: "acme", Suite: "students", Version: "v-view", Testcase: "bob",
				BuiltAt: time.Now().UTC(),
			},
			Results: []message.Result{{
				Key: "gpa", Kind: message.KindCheck,
				Value: message.Value{Type: message.TypeDouble, Double: &gpa},
			}},
		}},
	}
	body, _ := json.Marshal(env)
	req := httptest.NewRequest(http.MethodPost, "/v1/client/submit", bytes.NewReader(body))
	req.Header.Set("X-Recension-API-Key", viewerKey)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("viewer submit should be denied (404), got %d %s", rec.Code, rec.Body.String())
	}

	viewerTok := login(t, srv, "viewer@example.com", "password1")
	req = httptest.NewRequest(http.MethodGet, "/v1/teams/acme/suites", nil)
	req.Header.Set("Authorization", "Bearer "+viewerTok)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("viewer read: %d", rec.Code)
	}
}
