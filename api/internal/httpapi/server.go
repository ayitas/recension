package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ayitas/recension/api/internal/auth"
	"github.com/ayitas/recension/api/internal/blobstore"
	"github.com/ayitas/recension/api/internal/config"
	"github.com/ayitas/recension/api/internal/store"
	"github.com/ayitas/recension/api/internal/submit"
	"github.com/ayitas/recension/pkg/compare"
	"github.com/ayitas/recension/pkg/message"
)

type Server struct {
	store       store.Store
	config      config.Config
	tokens      *auth.Tokens
	mux         *http.ServeMux
	authLimiter *authLimiter
	blobs       *blobstore.Store
}

func New(st store.Store, cfg config.Config, blobs *blobstore.Store) *Server {
	s := &Server{
		store:       st,
		config:      cfg,
		tokens:      auth.NewTokens(cfg.SessionSecret, cfg.SessionTTL),
		mux:         http.NewServeMux(),
		authLimiter: newAuthLimiter(20, time.Minute),
		blobs:       blobs,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return securityHeaders(cors(s.config.CORSOrigins, s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)

	s.mux.HandleFunc("POST /v1/auth/signup", s.handleSignup)
	s.mux.HandleFunc("POST /v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /v1/auth/me", s.requireUser(s.handleMe))
	s.mux.HandleFunc("POST /v1/auth/api-key/rotate", s.requireUser(s.handleRotateAPIKey))

	s.mux.HandleFunc("POST /v1/client/verify", s.handleClientVerify)
	s.mux.HandleFunc("POST /v1/client/submit", s.handleClientSubmit)
	s.mux.HandleFunc("PUT /v1/blobs/{digest}", s.handlePutBlob)
	s.mux.HandleFunc("HEAD /v1/blobs/{digest}", s.handleHeadBlob)
	s.mux.HandleFunc("GET /v1/blobs/{digest}", s.handleGetBlob)
	s.mux.HandleFunc("POST /v1/batch/{team}/{suite}/{batch}/seal", s.handleBatchSeal)
	s.mux.HandleFunc("POST /v1/batch/{team}/{suite}/{batch}/promote", s.requireUser(s.requireTeamRole(store.RoleAdmin, s.handleBatchPromote)))

	s.mux.HandleFunc("GET /v1/teams", s.requireUser(s.handleListTeams))
	s.mux.HandleFunc("POST /v1/teams", s.requireUser(s.handleCreateTeam))
	s.mux.HandleFunc("GET /v1/teams/{team}/members", s.requireUser(s.requireTeamRole(store.RoleViewer, s.handleListMembers)))
	s.mux.HandleFunc("POST /v1/teams/{team}/members", s.requireUser(s.requireTeamRole(store.RoleAdmin, s.handleAddMember)))
	s.mux.HandleFunc("PATCH /v1/teams/{team}/members/{userId}", s.requireUser(s.requireTeamRole(store.RoleAdmin, s.handleUpdateMember)))
	s.mux.HandleFunc("DELETE /v1/teams/{team}/members/{userId}", s.requireUser(s.requireTeamRole(store.RoleAdmin, s.handleRemoveMember)))
	s.mux.HandleFunc("GET /v1/teams/{team}/suites", s.requireUser(s.requireTeamRole(store.RoleViewer, s.handleListSuites)))
	s.mux.HandleFunc("POST /v1/teams/{team}/suites", s.requireUser(s.requireTeamRole(store.RoleAdmin, s.handleCreateSuite)))
	s.mux.HandleFunc("GET /v1/teams/{team}/suites/{suite}/batches", s.requireUser(s.requireTeamRole(store.RoleViewer, s.handleListBatches)))
	s.mux.HandleFunc("GET /v1/teams/{team}/suites/{suite}/batches/{batch}", s.requireUser(s.requireTeamRole(store.RoleViewer, s.handleGetBatch)))
	s.mux.HandleFunc("GET /v1/teams/{team}/suites/{suite}/batches/{batch}/elements/{element}", s.requireUser(s.requireTeamRole(store.RoleViewer, s.handleGetElement)))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleClientVerify(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticateClient(r); !ok {
		writeErr(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleClientSubmit(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateClient(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxBodyBytes)
	var env message.Envelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) || errors.Is(err, io.ErrUnexpectedEOF) {
			writeErr(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	teamSlug := ""
	if len(env.Messages) > 0 {
		teamSlug = strings.TrimSpace(env.Messages[0].Metadata.Team)
	}
	if teamSlug == "" {
		writeErr(w, http.StatusBadRequest, "message metadata incomplete")
		return
	}
	if !s.authorizeTeamRole(w, user, teamSlug, store.RoleMember) {
		return
	}
	// All messages in one envelope must target the same authorized team.
	for _, msg := range env.Messages {
		if strings.TrimSpace(msg.Metadata.Team) != teamSlug {
			writeErr(w, http.StatusBadRequest, "all messages must target the same team")
			return
		}
	}
	outcomes, err := submit.Process(s.store, env)
	if err != nil {
		msg := err.Error()
		status := http.StatusBadRequest
		if strings.Contains(msg, "sealed") {
			status = http.StatusConflict
		}
		if strings.Contains(msg, "not found") {
			status = http.StatusNotFound
		}
		writeErr(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"outcomes": outcomes})
}

func (s *Server) handleBatchSeal(w http.ResponseWriter, r *http.Request) {
	user, ok := s.authenticateClient(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	team := r.PathValue("team")
	if !s.authorizeTeamRole(w, user, team, store.RoleMember) {
		return
	}
	suite := r.PathValue("suite")
	batchSlug := r.PathValue("batch")
	batch, _, ok := s.store.GetBatch(team, suite, batchSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "batch not found")
		return
	}
	s.store.SealBatch(batch)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleBatchPromote(w http.ResponseWriter, r *http.Request) {
	team := r.PathValue("team")
	suiteSlug := r.PathValue("suite")
	batchSlug := r.PathValue("batch")
	batch, suite, ok := s.store.GetBatch(team, suiteSlug, batchSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "batch not found")
		return
	}
	s.store.PromoteBaseline(suite, batch)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, s.store.ListTeamsForUser(user.ID))
}

func (s *Server) handleListSuites(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.ListSuites(r.PathValue("team")))
}

func (s *Server) handleListBatches(w http.ResponseWriter, r *http.Request) {
	team := r.PathValue("team")
	suiteSlug := r.PathValue("suite")
	batches := s.store.ListBatches(team, suiteSlug)

	var suite *store.Suite
	for _, su := range s.store.ListSuites(team) {
		if su.Slug == suiteSlug {
			su := su
			suite = &su
			break
		}
	}

	out := make([]map[string]any, 0, len(batches))
	for _, b := range batches {
		msgs := s.store.ListMessages(b.ID)
		cmps := s.store.ComparisonsForBatch(b.ID)
		byMsg := make(map[string]compare.Result, len(cmps))
		for _, c := range cmps {
			byMsg[c.SrcMessageID] = c.Result
		}

		isBaseline := suite != nil && suite.BaselineBatchID == b.ID
		pass, diff, sent := 0, 0, 0
		var scoreSum float64
		for _, msg := range msgs {
			if c, ok := byMsg[msg.ID]; ok {
				switch c.Verdict() {
				case "diff":
					diff++
				case "sent":
					sent++
				default:
					pass++
				}
				scoreSum += c.Overview.KeysScore
				continue
			}
			if isBaseline {
				sent++
			} else {
				pass++
			}
			scoreSum += 1
		}

		n := len(msgs)
		avg := 1.0
		if n > 0 {
			avg = scoreSum / float64(n)
		}

		item := map[string]any{
			"id":            b.ID,
			"suiteId":       b.SuiteID,
			"slug":          b.Slug,
			"submittedAt":   b.SubmittedAt,
			"meta":          b.Meta,
			"testcaseCount": n,
			"passCount":     pass,
			"diffCount":     diff,
			"sentCount":     sent,
			"avgScore":      avg,
			"isBaseline":    isBaseline,
		}
		if b.SealedAt != nil {
			item["sealedAt"] = b.SealedAt
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	team := r.PathValue("team")
	suiteSlug := r.PathValue("suite")
	batchSlug := r.PathValue("batch")
	batch, suite, ok := s.store.GetBatch(team, suiteSlug, batchSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "batch not found")
		return
	}
	msgs := s.store.ListMessages(batch.ID)
	cmps := s.store.ComparisonsForBatch(batch.ID)
	elements := make([]map[string]any, 0, len(msgs))
	for _, msg := range msgs {
		el, ok := s.store.ElementByID(msg.ElementID)
		name := msg.ElementID
		if ok {
			name = el.Name
		}
		item := map[string]any{
			"testcase": name,
			"builtAt":  msg.BuiltAt,
		}
		for _, c := range cmps {
			if c.SrcMessageID == msg.ID {
				item["verdict"] = c.Result.Verdict()
				item["score"] = c.Result.Overview.KeysScore
				item["comparison"] = c.Result
				break
			}
		}
		if _, has := item["verdict"]; !has {
			if suite.BaselineBatchID == batch.ID {
				item["verdict"] = "sent"
				item["score"] = 1.0
			} else {
				item["verdict"] = "pass"
				item["score"] = 1.0
			}
		}
		elements = append(elements, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"batch":           batch,
		"baselineBatchId": suite.BaselineBatchID,
		"elements":        elements,
	})
}

func (s *Server) handleGetElement(w http.ResponseWriter, r *http.Request) {
	team := r.PathValue("team")
	suiteSlug := r.PathValue("suite")
	batchSlug := r.PathValue("batch")
	elementName := r.PathValue("element")

	batch, suite, ok := s.store.GetBatch(team, suiteSlug, batchSlug)
	if !ok {
		writeErr(w, http.StatusNotFound, "batch not found")
		return
	}

	var msg *store.MessageRecord
	var elName string
	for _, m := range s.store.ListMessages(batch.ID) {
		el, found := s.store.ElementByID(m.ElementID)
		name := m.ElementID
		if found {
			name = el.Name
		}
		if name == elementName {
			rec := m
			msg = &rec
			elName = name
			break
		}
	}
	if msg == nil {
		writeErr(w, http.StatusNotFound, "element not found")
		return
	}

	verdict := "pass"
	score := 1.0
	var cmp *compare.Result

	for _, c := range s.store.ComparisonsForBatch(batch.ID) {
		if c.SrcMessageID == msg.ID {
			result := c.Result
			cmp = &result
			verdict = result.Verdict()
			score = result.Overview.KeysScore
			break
		}
	}

	if cmp == nil {
		if suite.BaselineBatchID == batch.ID {
			self := compare.Messages(msg.Payload, msg.Payload)
			cmp = &self
			verdict = "sent"
			score = 1.0
		} else if baseline, ok := s.store.BaselineBatch(suite); ok {
			if dst, ok := s.store.MessageByBatchElement(baseline.ID, msg.ElementID); ok {
				result := compare.Messages(msg.Payload, dst.Payload)
				cmp = &result
				verdict = result.Verdict()
				score = result.Overview.KeysScore
			} else {
				self := compare.Messages(msg.Payload, msg.Payload)
				cmp = &self
				verdict = "pass"
				score = 1.0
			}
		} else {
			self := compare.Messages(msg.Payload, msg.Payload)
			cmp = &self
			verdict = "sent"
			score = 1.0
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"testcase":        elName,
		"verdict":         verdict,
		"score":           score,
		"batch":           batch,
		"baselineBatchId": suite.BaselineBatchID,
		"message":         msg.Payload,
		"comparison":      cmp,
	})
}

func (s *Server) authenticateClient(r *http.Request) (*store.User, bool) {
	key := r.Header.Get("X-Recension-API-Key")
	if key == "" {
		return nil, false
	}
	return s.store.UserByAPIKey(key)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"errors": []string{msg}})
}

func cors(origins []string, next http.Handler) http.Handler {
	allowAll := len(origins) == 1 && origins[0] == "*"
	allowed := map[string]struct{}{}
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Recension-API-Key, X-Recension-Mime")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
