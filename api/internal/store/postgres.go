package store

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ayitas/recension/pkg/message"
)

//go:embed migrate.sql
var migrateFS embed.FS

// Postgres implements Store using PostgreSQL.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, databaseURL, bootstrapAPIKey, bootstrapPasswordHash string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	p := &Postgres{pool: pool}
	if err := p.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := p.bootstrap(ctx, bootstrapAPIKey, bootstrapPasswordHash); err != nil {
		pool.Close()
		return nil, err
	}
	return p, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) migrate(ctx context.Context) error {
	sqlBytes, err := migrateFS.ReadFile("migrate.sql")
	if err != nil {
		return err
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`,
		"001_init",
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) bootstrap(ctx context.Context, apiKey, passwordHash string) error {
	if apiKey == "" || passwordHash == "" {
		return nil
	}

	var exists bool
	err := p.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		"dev@recension.local").Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		_, err = p.pool.Exec(ctx,
			`INSERT INTO users (id, email, api_key, password_hash) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (email) DO NOTHING`,
			uuid.NewString(), "dev@recension.local", apiKey, passwordHash,
		)
		if err != nil {
			return err
		}
	}

	var teamID string
	err = p.pool.QueryRow(ctx, `SELECT id::text FROM teams WHERE slug = 'acme'`).Scan(&teamID)
	if err != nil {
		teamID = uuid.NewString()
		_, err = p.pool.Exec(ctx,
			`INSERT INTO teams (id, slug, name) VALUES ($1, 'acme', 'Acme')
			 ON CONFLICT (slug) DO NOTHING`,
			teamID,
		)
		if err != nil {
			return err
		}
		_ = p.pool.QueryRow(ctx, `SELECT id::text FROM teams WHERE slug = 'acme'`).Scan(&teamID)
	}

	var suiteExists bool
	err = p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM suites WHERE team_id = $1::uuid AND slug = 'students')`,
		teamID,
	).Scan(&suiteExists)
	if err != nil {
		return err
	}
	if !suiteExists {
		_, err = p.pool.Exec(ctx,
			`INSERT INTO suites (id, team_id, slug, name) VALUES ($1, $2::uuid, 'students', 'Students')
			 ON CONFLICT (team_id, slug) DO NOTHING`,
			uuid.NewString(), teamID,
		)
		if err != nil {
			return err
		}
	}

	var artifactsExists bool
	err = p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM suites WHERE team_id = $1::uuid AND slug = 'artifacts')`,
		teamID,
	).Scan(&artifactsExists)
	if err != nil {
		return err
	}
	if !artifactsExists {
		_, err = p.pool.Exec(ctx,
			`INSERT INTO suites (id, team_id, slug, name) VALUES ($1, $2::uuid, 'artifacts', 'Artifacts')
			 ON CONFLICT (team_id, slug) DO NOTHING`,
			uuid.NewString(), teamID,
		)
		if err != nil {
			return err
		}
	}

	var exportsExists bool
	err = p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM suites WHERE team_id = $1::uuid AND slug = 'exports')`,
		teamID,
	).Scan(&exportsExists)
	if err != nil {
		return err
	}
	if !exportsExists {
		_, err = p.pool.Exec(ctx,
			`INSERT INTO suites (id, team_id, slug, name) VALUES ($1, $2::uuid, 'exports', 'Exports')
			 ON CONFLICT (team_id, slug) DO NOTHING`,
			uuid.NewString(), teamID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) UserByAPIKey(key string) (*User, bool) {
	ctx := context.Background()
	var u User
	var hash *string
	err := p.pool.QueryRow(ctx,
		`SELECT id::text, email, api_key, password_hash FROM users WHERE api_key = $1`, key,
	).Scan(&u.ID, &u.Email, &u.APIKey, &hash)
	if err != nil {
		return nil, false
	}
	if hash != nil {
		u.PasswordHash = *hash
	}
	return &u, true
}

func (p *Postgres) UserByEmail(email string) (*User, bool) {
	ctx := context.Background()
	var u User
	var hash *string
	err := p.pool.QueryRow(ctx,
		`SELECT id::text, email, api_key, password_hash FROM users WHERE lower(email) = lower($1)`, email,
	).Scan(&u.ID, &u.Email, &u.APIKey, &hash)
	if err != nil {
		return nil, false
	}
	if hash != nil {
		u.PasswordHash = *hash
	}
	return &u, true
}

func (p *Postgres) UserByID(id string) (*User, bool) {
	ctx := context.Background()
	var u User
	var hash *string
	err := p.pool.QueryRow(ctx,
		`SELECT id::text, email, api_key, password_hash FROM users WHERE id = $1::uuid`, id,
	).Scan(&u.ID, &u.Email, &u.APIKey, &hash)
	if err != nil {
		return nil, false
	}
	if hash != nil {
		u.PasswordHash = *hash
	}
	return &u, true
}

func (p *Postgres) CreateUser(email, passwordHash, apiKey string) (*User, error) {
	ctx := context.Background()
	u := &User{
		ID:           uuid.NewString(),
		Email:        email,
		APIKey:       apiKey,
		PasswordHash: passwordHash,
	}
	_, err := p.pool.Exec(ctx,
		`INSERT INTO users (id, email, api_key, password_hash) VALUES ($1, $2, $3, $4)`,
		u.ID, u.Email, u.APIKey, u.PasswordHash,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) RotateAPIKey(userID, newKey string) (*User, error) {
	ctx := context.Background()
	var u User
	var hash *string
	err := p.pool.QueryRow(ctx, `
		UPDATE users SET api_key = $1 WHERE id = $2::uuid
		RETURNING id::text, email, api_key, password_hash`,
		newKey, userID,
	).Scan(&u.ID, &u.Email, &u.APIKey, &hash)
	if err != nil {
		return nil, err
	}
	if hash != nil {
		u.PasswordHash = *hash
	}
	return &u, nil
}

func (p *Postgres) ListTeams() []Team {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, `SELECT id::text, slug, name FROM teams ORDER BY slug`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []Team{}
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (p *Postgres) EnsureTeam(slug, name string) *Team {
	ctx := context.Background()
	if name == "" {
		name = slug
	}
	var t Team
	err := p.pool.QueryRow(ctx,
		`INSERT INTO teams (id, slug, name) VALUES ($1, $2, $3)
		 ON CONFLICT (slug) DO UPDATE SET name = COALESCE(NULLIF(EXCLUDED.name, ''), teams.name)
		 RETURNING id::text, slug, name`,
		uuid.NewString(), slug, name,
	).Scan(&t.ID, &t.Slug, &t.Name)
	if err != nil {
		_ = p.pool.QueryRow(ctx, `SELECT id::text, slug, name FROM teams WHERE slug = $1`, slug).
			Scan(&t.ID, &t.Slug, &t.Name)
	}
	return &t
}

func (p *Postgres) EnsureSuite(teamSlug, suiteSlug, name string) (*Suite, error) {
	ctx := context.Background()
	var teamID string
	err := p.pool.QueryRow(ctx, `SELECT id::text FROM teams WHERE slug = $1`, teamSlug).Scan(&teamID)
	if err != nil {
		return nil, fmt.Errorf("team %q not found", teamSlug)
	}
	if name == "" {
		name = suiteSlug
	}
	var s Suite
	var baseline *string
	err = p.pool.QueryRow(ctx,
		`INSERT INTO suites (id, team_id, slug, name)
		 VALUES ($1, $2::uuid, $3, $4)
		 ON CONFLICT (team_id, slug) DO UPDATE SET name = COALESCE(NULLIF(EXCLUDED.name, ''), suites.name)
		 RETURNING id::text, team_id::text, slug, name, baseline_batch_id::text`,
		uuid.NewString(), teamID, suiteSlug, name,
	).Scan(&s.ID, &s.TeamID, &s.Slug, &s.Name, &baseline)
	if err != nil {
		return nil, err
	}
	if baseline != nil {
		s.BaselineBatchID = *baseline
	}
	return &s, nil
}

func (p *Postgres) ListSuites(teamSlug string) []Suite {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, `
		SELECT s.id::text, s.team_id::text, s.slug, s.name, COALESCE(s.baseline_batch_id::text, '')
		FROM suites s
		JOIN teams t ON t.id = s.team_id
		WHERE t.slug = $1
		ORDER BY s.slug`, teamSlug)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []Suite{}
	for rows.Next() {
		var s Suite
		if err := rows.Scan(&s.ID, &s.TeamID, &s.Slug, &s.Name, &s.BaselineBatchID); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (p *Postgres) EnsureBatch(suite *Suite, slug string) *Batch {
	ctx := context.Background()
	var b Batch
	var sealed *time.Time
	var meta []byte
	err := p.pool.QueryRow(ctx,
		`INSERT INTO batches (id, suite_id, slug, submitted_at, meta)
		 VALUES ($1, $2::uuid, $3, $4, '{}'::jsonb)
		 ON CONFLICT (suite_id, slug) DO UPDATE SET slug = EXCLUDED.slug
		 RETURNING id::text, suite_id::text, slug, sealed_at, submitted_at, meta`,
		uuid.NewString(), suite.ID, slug, time.Now().UTC(),
	).Scan(&b.ID, &b.SuiteID, &b.Slug, &sealed, &b.SubmittedAt, &meta)
	if err != nil {
		return &Batch{ID: uuid.NewString(), SuiteID: suite.ID, Slug: slug, SubmittedAt: time.Now().UTC(), Meta: map[string]any{}}
	}
	b.SealedAt = sealed
	b.Meta = decodeMeta(meta)
	return &b
}

func (p *Postgres) ListBatches(teamSlug, suiteSlug string) []Batch {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, `
		SELECT b.id::text, b.suite_id::text, b.slug, b.sealed_at, b.submitted_at, b.meta
		FROM batches b
		JOIN suites s ON s.id = b.suite_id
		JOIN teams t ON t.id = s.team_id
		WHERE t.slug = $1 AND s.slug = $2
		ORDER BY b.submitted_at DESC`, teamSlug, suiteSlug)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanBatches(rows)
}

func (p *Postgres) GetBatch(teamSlug, suiteSlug, batchSlug string) (*Batch, *Suite, bool) {
	ctx := context.Background()
	var s Suite
	var baseline *string
	err := p.pool.QueryRow(ctx, `
		SELECT s.id::text, s.team_id::text, s.slug, s.name, s.baseline_batch_id::text
		FROM suites s
		JOIN teams t ON t.id = s.team_id
		WHERE t.slug = $1 AND s.slug = $2`, teamSlug, suiteSlug,
	).Scan(&s.ID, &s.TeamID, &s.Slug, &s.Name, &baseline)
	if err != nil {
		return nil, nil, false
	}
	if baseline != nil {
		s.BaselineBatchID = *baseline
	}

	var b Batch
	var sealed *time.Time
	var meta []byte
	err = p.pool.QueryRow(ctx, `
		SELECT id::text, suite_id::text, slug, sealed_at, submitted_at, meta
		FROM batches WHERE suite_id = $1::uuid AND slug = $2`, s.ID, batchSlug,
	).Scan(&b.ID, &b.SuiteID, &b.Slug, &sealed, &b.SubmittedAt, &meta)
	if err != nil {
		return nil, nil, false
	}
	b.SealedAt = sealed
	b.Meta = decodeMeta(meta)
	return &b, &s, true
}

func (p *Postgres) EnsureElement(suite *Suite, name string) *Element {
	ctx := context.Background()
	var e Element
	err := p.pool.QueryRow(ctx,
		`INSERT INTO elements (id, suite_id, name)
		 VALUES ($1, $2::uuid, $3)
		 ON CONFLICT (suite_id, name) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id::text, suite_id::text, name`,
		uuid.NewString(), suite.ID, name,
	).Scan(&e.ID, &e.SuiteID, &e.Name)
	if err != nil {
		_ = p.pool.QueryRow(ctx,
			`SELECT id::text, suite_id::text, name FROM elements WHERE suite_id = $1::uuid AND name = $2`,
			suite.ID, name,
		).Scan(&e.ID, &e.SuiteID, &e.Name)
	}
	return &e
}

func (p *Postgres) PutMessage(batch *Batch, element *Element, msg message.Message) *MessageRecord {
	ctx := context.Background()
	payload, err := json.Marshal(msg)
	if err != nil {
		payload = []byte("{}")
	}
	id := uuid.NewString()
	builtAt := msg.Metadata.BuiltAt
	if builtAt.IsZero() {
		builtAt = time.Now().UTC()
	}
	var rec MessageRecord
	var raw []byte
	err = p.pool.QueryRow(ctx, `
		INSERT INTO messages (id, batch_id, element_id, built_at, payload)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5::jsonb)
		ON CONFLICT (batch_id, element_id) DO UPDATE
		  SET built_at = EXCLUDED.built_at,
		      payload = EXCLUDED.payload,
		      id = messages.id
		RETURNING id::text, batch_id::text, element_id::text, built_at, payload`,
		id, batch.ID, element.ID, builtAt, payload,
	).Scan(&rec.ID, &rec.BatchID, &rec.ElementID, &rec.BuiltAt, &raw)
	if err != nil {
		rec = MessageRecord{ID: id, BatchID: batch.ID, ElementID: element.ID, BuiltAt: builtAt, Payload: msg}
		return &rec
	}
	_ = json.Unmarshal(raw, &rec.Payload)
	return &rec
}

func (p *Postgres) MessageByBatchElement(batchID, elementID string) (*MessageRecord, bool) {
	ctx := context.Background()
	var rec MessageRecord
	var raw []byte
	err := p.pool.QueryRow(ctx, `
		SELECT id::text, batch_id::text, element_id::text, built_at, payload
		FROM messages WHERE batch_id = $1::uuid AND element_id = $2::uuid`,
		batchID, elementID,
	).Scan(&rec.ID, &rec.BatchID, &rec.ElementID, &rec.BuiltAt, &raw)
	if err != nil {
		return nil, false
	}
	_ = json.Unmarshal(raw, &rec.Payload)
	return &rec, true
}

func (p *Postgres) ListMessages(batchID string) []MessageRecord {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, `
		SELECT id::text, batch_id::text, element_id::text, built_at, payload
		FROM messages WHERE batch_id = $1::uuid ORDER BY built_at`, batchID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []MessageRecord{}
	for rows.Next() {
		var rec MessageRecord
		var raw []byte
		if err := rows.Scan(&rec.ID, &rec.BatchID, &rec.ElementID, &rec.BuiltAt, &raw); err != nil {
			continue
		}
		_ = json.Unmarshal(raw, &rec.Payload)
		out = append(out, rec)
	}
	return out
}

func (p *Postgres) SaveComparison(rec ComparisonRecord) *ComparisonRecord {
	ctx := context.Background()
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	raw, err := json.Marshal(rec.Result)
	if err != nil {
		raw = []byte("{}")
	}
	// Replace prior comparison for same src message.
	_, _ = p.pool.Exec(ctx, `DELETE FROM comparisons WHERE src_message_id = $1::uuid`, rec.SrcMessageID)
	_, err = p.pool.Exec(ctx, `
		INSERT INTO comparisons (id, src_message_id, dst_message_id, src_batch_id, dst_batch_id, result)
		VALUES ($1, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::jsonb)`,
		rec.ID, rec.SrcMessageID, rec.DstMessageID, rec.SrcBatchID, rec.DstBatchID, raw,
	)
	if err != nil {
		return &rec
	}
	return &rec
}

func (p *Postgres) ComparisonsForBatch(batchID string) []ComparisonRecord {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, `
		SELECT id::text, src_message_id::text, dst_message_id::text,
		       src_batch_id::text, dst_batch_id::text, result
		FROM comparisons WHERE src_batch_id = $1::uuid`, batchID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []ComparisonRecord{}
	for rows.Next() {
		var c ComparisonRecord
		var raw []byte
		if err := rows.Scan(&c.ID, &c.SrcMessageID, &c.DstMessageID, &c.SrcBatchID, &c.DstBatchID, &raw); err != nil {
			continue
		}
		_ = json.Unmarshal(raw, &c.Result)
		out = append(out, c)
	}
	return out
}

func (p *Postgres) SealBatch(batch *Batch) {
	ctx := context.Background()
	now := time.Now().UTC()
	_, _ = p.pool.Exec(ctx, `UPDATE batches SET sealed_at = $1 WHERE id = $2::uuid`, now, batch.ID)
	batch.SealedAt = &now
}

func (p *Postgres) PromoteBaseline(suite *Suite, batch *Batch) {
	ctx := context.Background()
	_, _ = p.pool.Exec(ctx, `UPDATE suites SET baseline_batch_id = $1::uuid WHERE id = $2::uuid`, batch.ID, suite.ID)
	suite.BaselineBatchID = batch.ID
}

func (p *Postgres) BaselineBatch(suite *Suite) (*Batch, bool) {
	if suite.BaselineBatchID == "" {
		// reload from DB in case caller has stale suite
		ctx := context.Background()
		var baseline *string
		_ = p.pool.QueryRow(ctx, `SELECT baseline_batch_id::text FROM suites WHERE id = $1::uuid`, suite.ID).Scan(&baseline)
		if baseline == nil || *baseline == "" {
			return nil, false
		}
		suite.BaselineBatchID = *baseline
	}
	ctx := context.Background()
	var b Batch
	var sealed *time.Time
	var meta []byte
	err := p.pool.QueryRow(ctx, `
		SELECT id::text, suite_id::text, slug, sealed_at, submitted_at, meta
		FROM batches WHERE id = $1::uuid`, suite.BaselineBatchID,
	).Scan(&b.ID, &b.SuiteID, &b.Slug, &sealed, &b.SubmittedAt, &meta)
	if err != nil {
		return nil, false
	}
	b.SealedAt = sealed
	b.Meta = decodeMeta(meta)
	return &b, true
}

func (p *Postgres) ElementByID(id string) (*Element, bool) {
	ctx := context.Background()
	var e Element
	err := p.pool.QueryRow(ctx,
		`SELECT id::text, suite_id::text, name FROM elements WHERE id = $1::uuid`, id,
	).Scan(&e.ID, &e.SuiteID, &e.Name)
	if err != nil {
		return nil, false
	}
	return &e, true
}

func (p *Postgres) MessageByID(id string) (*MessageRecord, bool) {
	ctx := context.Background()
	var rec MessageRecord
	var raw []byte
	err := p.pool.QueryRow(ctx, `
		SELECT id::text, batch_id::text, element_id::text, built_at, payload
		FROM messages WHERE id = $1::uuid`, id,
	).Scan(&rec.ID, &rec.BatchID, &rec.ElementID, &rec.BuiltAt, &raw)
	if err != nil {
		return nil, false
	}
	_ = json.Unmarshal(raw, &rec.Payload)
	return &rec, true
}

func decodeMeta(raw []byte) map[string]any {
	meta := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &meta)
	}
	return meta
}

func scanBatches(rows pgx.Rows) []Batch {
	out := []Batch{}
	for rows.Next() {
		var b Batch
		var sealed *time.Time
		var meta []byte
		if err := rows.Scan(&b.ID, &b.SuiteID, &b.Slug, &sealed, &b.SubmittedAt, &meta); err != nil {
			continue
		}
		b.SealedAt = sealed
		b.Meta = decodeMeta(meta)
		out = append(out, b)
	}
	return out
}