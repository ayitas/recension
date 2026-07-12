package store

import (
	"time"

	"github.com/ayitas/recension/pkg/compare"
	"github.com/ayitas/recension/pkg/message"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	APIKey       string `json:"apiKey"`
	PasswordHash string `json:"-"`
}

type Team struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Suite struct {
	ID              string `json:"id"`
	TeamID          string `json:"teamId"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	BaselineBatchID string `json:"baselineBatchId"`
}

type Batch struct {
	ID          string         `json:"id"`
	SuiteID     string         `json:"suiteId"`
	Slug        string         `json:"slug"`
	SealedAt    *time.Time     `json:"sealedAt,omitempty"`
	SubmittedAt time.Time      `json:"submittedAt"`
	Meta        map[string]any `json:"meta"`
}

type Element struct {
	ID      string `json:"id"`
	SuiteID string `json:"suiteId"`
	Name    string `json:"name"`
}

type MessageRecord struct {
	ID        string
	BatchID   string
	ElementID string
	BuiltAt   time.Time
	Payload   message.Message
}

type ComparisonRecord struct {
	ID           string
	SrcMessageID string
	DstMessageID string
	SrcBatchID   string
	DstBatchID   string
	Result       compare.Result
}

// Store is the persistence API used by HTTP handlers and submit flow.
type Store interface {
	UserByAPIKey(key string) (*User, bool)
	UserByEmail(email string) (*User, bool)
	UserByID(id string) (*User, bool)
	CreateUser(email, passwordHash, apiKey string) (*User, error)
	RotateAPIKey(userID, newKey string) (*User, error)

	TeamBySlug(slug string) (*Team, bool)
	ListTeamsForUser(userID string) []Team
	EnsureTeam(slug, name string) *Team
	Membership(userID, teamSlug string) (role string, ok bool)
	AddMember(teamID, userID, role string) error
	RemoveMember(teamID, userID string) error
	SetMemberRole(teamID, userID, role string) error
	ListMembers(teamSlug string) ([]TeamMember, error)
	CountOwners(teamID string) int

	CreateInvite(teamID, createdBy, role string, ttl time.Duration) (*TeamInvite, error)
	InviteByToken(token string) (*TeamInvite, *Team, error)
	AcceptInvite(token, userID string) (*TeamMember, error)

	EnsureSuite(teamSlug, suiteSlug, name string) (*Suite, error)
	ListSuites(teamSlug string) []Suite
	EnsureBatch(suite *Suite, slug string) *Batch
	ListBatches(teamSlug, suiteSlug string) []Batch
	GetBatch(teamSlug, suiteSlug, batchSlug string) (*Batch, *Suite, bool)
	EnsureElement(suite *Suite, name string) *Element
	PutMessage(batch *Batch, element *Element, msg message.Message) *MessageRecord
	MessageByBatchElement(batchID, elementID string) (*MessageRecord, bool)
	ListMessages(batchID string) []MessageRecord
	SaveComparison(rec ComparisonRecord) *ComparisonRecord
	ComparisonsForBatch(batchID string) []ComparisonRecord
	SealBatch(batch *Batch)
	PromoteBaseline(suite *Suite, batch *Batch)
	BaselineBatch(suite *Suite) (*Batch, bool)
	ElementByID(id string) (*Element, bool)
	MessageByID(id string) (*MessageRecord, bool)
}
