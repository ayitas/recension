package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ayitas/recension/pkg/message"
)

// Memory is an in-memory store for tests / offline demos.
type Memory struct {
	mu sync.RWMutex

	users       map[string]*User
	usersByID   map[string]*User
	teams       map[string]*Team
	suites      map[string]*Suite
	batches     map[string]*Batch
	elements    map[string]*Element
	messages    map[string]*MessageRecord
	msgByBE     map[string]string
	comparisons map[string]*ComparisonRecord
}

func NewMemory(bootstrapAPIKey, bootstrapPasswordHash string) *Memory {
	m := &Memory{
		users:       map[string]*User{},
		usersByID:   map[string]*User{},
		teams:       map[string]*Team{},
		suites:      map[string]*Suite{},
		batches:     map[string]*Batch{},
		elements:    map[string]*Element{},
		messages:    map[string]*MessageRecord{},
		msgByBE:     map[string]string{},
		comparisons: map[string]*ComparisonRecord{},
	}
	u := &User{
		ID:           uuid.NewString(),
		Email:        "dev@recension.local",
		APIKey:       bootstrapAPIKey,
		PasswordHash: bootstrapPasswordHash,
	}
	m.users[u.APIKey] = u
	m.usersByID[u.ID] = u

	team := &Team{ID: uuid.NewString(), Slug: "acme", Name: "Acme"}
	m.teams[team.Slug] = team
	suite := &Suite{ID: uuid.NewString(), TeamID: team.ID, Slug: "students", Name: "Students"}
	m.suites[team.Slug+"/"+suite.Slug] = suite
	artifacts := &Suite{ID: uuid.NewString(), TeamID: team.ID, Slug: "artifacts", Name: "Artifacts"}
	m.suites[team.Slug+"/"+artifacts.Slug] = artifacts
	exports := &Suite{ID: uuid.NewString(), TeamID: team.ID, Slug: "exports", Name: "Exports"}
	m.suites[team.Slug+"/"+exports.Slug] = exports
	return m
}

func (m *Memory) UserByAPIKey(key string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[key]
	return u, ok
}

func (m *Memory) UserByEmail(email string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.usersByID {
		if strings.EqualFold(u.Email, email) {
			return u, true
		}
	}
	return nil, false
}

func (m *Memory) UserByID(id string) (*User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.usersByID[id]
	return u, ok
}

func (m *Memory) CreateUser(email, passwordHash, apiKey string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.usersByID {
		if strings.EqualFold(u.Email, email) {
			return nil, fmt.Errorf("email already registered")
		}
	}
	u := &User{
		ID:           uuid.NewString(),
		Email:        email,
		APIKey:       apiKey,
		PasswordHash: passwordHash,
	}
	m.users[u.APIKey] = u
	m.usersByID[u.ID] = u
	return u, nil
}

func (m *Memory) RotateAPIKey(userID, newKey string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.usersByID[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	delete(m.users, u.APIKey)
	u.APIKey = newKey
	m.users[newKey] = u
	return u, nil
}

func (m *Memory) ListTeams() []Team {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Team, 0, len(m.teams))
	for _, t := range m.teams {
		out = append(out, *t)
	}
	return out
}

func (m *Memory) EnsureTeam(slug, name string) *Team {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.teams[slug]; ok {
		return t
	}
	if name == "" {
		name = slug
	}
	t := &Team{ID: uuid.NewString(), Slug: slug, Name: name}
	m.teams[slug] = t
	return t
}

func (m *Memory) EnsureSuite(teamSlug, suiteSlug, name string) (*Suite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	team, ok := m.teams[teamSlug]
	if !ok {
		return nil, fmt.Errorf("team %q not found", teamSlug)
	}
	key := teamSlug + "/" + suiteSlug
	if s, ok := m.suites[key]; ok {
		return s, nil
	}
	if name == "" {
		name = suiteSlug
	}
	s := &Suite{ID: uuid.NewString(), TeamID: team.ID, Slug: suiteSlug, Name: name}
	m.suites[key] = s
	return s, nil
}

func (m *Memory) ListSuites(teamSlug string) []Suite {
	m.mu.RLock()
	defer m.mu.RUnlock()
	team, ok := m.teams[teamSlug]
	if !ok {
		return nil
	}
	out := []Suite{}
	for _, s := range m.suites {
		if s.TeamID == team.ID {
			out = append(out, *s)
		}
	}
	return out
}

func (m *Memory) EnsureBatch(suite *Suite, slug string) *Batch {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := suite.ID + "/" + slug
	if b, ok := m.batches[key]; ok {
		return b
	}
	b := &Batch{
		ID:          uuid.NewString(),
		SuiteID:     suite.ID,
		Slug:        slug,
		SubmittedAt: time.Now().UTC(),
		Meta:        map[string]any{},
	}
	m.batches[key] = b
	return b
}

func (m *Memory) ListBatches(teamSlug, suiteSlug string) []Batch {
	m.mu.RLock()
	defer m.mu.RUnlock()
	suite, ok := m.suites[teamSlug+"/"+suiteSlug]
	if !ok {
		return nil
	}
	out := []Batch{}
	for _, b := range m.batches {
		if b.SuiteID == suite.ID {
			out = append(out, *b)
		}
	}
	return out
}

func (m *Memory) GetBatch(teamSlug, suiteSlug, batchSlug string) (*Batch, *Suite, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	suite, ok := m.suites[teamSlug+"/"+suiteSlug]
	if !ok {
		return nil, nil, false
	}
	b, ok := m.batches[suite.ID+"/"+batchSlug]
	return b, suite, ok
}

func (m *Memory) EnsureElement(suite *Suite, name string) *Element {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := suite.ID + "/" + name
	if e, ok := m.elements[key]; ok {
		return e
	}
	e := &Element{ID: uuid.NewString(), SuiteID: suite.ID, Name: name}
	m.elements[key] = e
	return e
}

func (m *Memory) PutMessage(batch *Batch, element *Element, msg message.Message) *MessageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := &MessageRecord{
		ID:        uuid.NewString(),
		BatchID:   batch.ID,
		ElementID: element.ID,
		BuiltAt:   msg.Metadata.BuiltAt,
		Payload:   msg,
	}
	m.messages[rec.ID] = rec
	m.msgByBE[batch.ID+"|"+element.ID] = rec.ID
	return rec
}

func (m *Memory) MessageByBatchElement(batchID, elementID string) (*MessageRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.msgByBE[batchID+"|"+elementID]
	if !ok {
		return nil, false
	}
	rec, ok := m.messages[id]
	return rec, ok
}

func (m *Memory) ListMessages(batchID string) []MessageRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []MessageRecord{}
	for _, rec := range m.messages {
		if rec.BatchID == batchID {
			out = append(out, *rec)
		}
	}
	return out
}

func (m *Memory) SaveComparison(rec ComparisonRecord) *ComparisonRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	cp := rec
	m.comparisons[cp.ID] = &cp
	return &cp
}

func (m *Memory) ComparisonsForBatch(batchID string) []ComparisonRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []ComparisonRecord{}
	for _, c := range m.comparisons {
		if c.SrcBatchID == batchID {
			out = append(out, *c)
		}
	}
	return out
}

func (m *Memory) SealBatch(batch *Batch) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	batch.SealedAt = &now
}

func (m *Memory) PromoteBaseline(suite *Suite, batch *Batch) {
	m.mu.Lock()
	defer m.mu.Unlock()
	suite.BaselineBatchID = batch.ID
}

func (m *Memory) BaselineBatch(suite *Suite) (*Batch, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if suite.BaselineBatchID == "" {
		return nil, false
	}
	for _, b := range m.batches {
		if b.ID == suite.BaselineBatchID {
			return b, true
		}
	}
	return nil, false
}

func (m *Memory) ElementByID(id string) (*Element, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.elements {
		if e.ID == id {
			return e, true
		}
	}
	return nil, false
}

func (m *Memory) MessageByID(id string) (*MessageRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.messages[id]
	return rec, ok
}
