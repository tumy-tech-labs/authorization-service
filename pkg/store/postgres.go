package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/lib/pq"

	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/tenant"
)

// PostgresStore implements Store backed by a PostgreSQL database.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgres creates a new PostgresStore using the provided DSN.
func NewPostgres(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) SaveTenant(ctx context.Context, t tenant.Tenant) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tenants(id, name, created_at) VALUES($1,$2,$3)
         ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name, created_at=EXCLUDED.created_at`,
		t.ID, t.Name, t.CreatedAt.Unix())
	return err
}

func (s *PostgresStore) LoadTenant(ctx context.Context, id string) (tenant.Tenant, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM tenants WHERE id=$1`, id)
	var t tenant.Tenant
	var created int64
	if err := row.Scan(&t.ID, &t.Name, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenant.Tenant{}, errors.New("tenant not found")
		}
		return tenant.Tenant{}, err
	}
	t.CreatedAt = time.Unix(created, 0).UTC()
	return t, nil
}

func (s *PostgresStore) ListTenants(ctx context.Context) ([]tenant.Tenant, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at FROM tenants`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []tenant.Tenant{}
	for rows.Next() {
		var t tenant.Tenant
		var created int64
		if err := rows.Scan(&t.ID, &t.Name, &created); err != nil {
			return nil, err
		}
		t.CreatedAt = time.Unix(created, 0).UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *PostgresStore) DeleteTenant(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM tenants WHERE id=$1`, id); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM policies WHERE tenant_id=$1`, id); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM edges WHERE tenant_id=$1`, id)
	return err
}

func (s *PostgresStore) SavePolicy(ctx context.Context, tenantID string, p policy.Policy) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO policies(tenant_id, policy_id, policy) VALUES($1,$2,$3)
         ON CONFLICT(tenant_id, policy_id) DO UPDATE SET policy=EXCLUDED.policy`,
		tenantID, p.ID, string(b))
	return err
}

func (s *PostgresStore) LoadPolicies(ctx context.Context, tenantID string) ([]policy.Policy, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT policy FROM policies WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []policy.Policy{}
	for rows.Next() {
		var js string
		if err := rows.Scan(&js); err != nil {
			return nil, err
		}
		var p policy.Policy
		if err := json.Unmarshal([]byte(js), &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ClearPolicies(ctx context.Context, tenantID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM policies WHERE tenant_id=$1`, tenantID)
	return err
}

func (s *PostgresStore) SaveEdge(ctx context.Context, tenantID, src, dst string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO edges(tenant_id, src, dst) VALUES($1,$2,$3)
         ON CONFLICT DO NOTHING`,
		tenantID, src, dst)
	return err
}

func (s *PostgresStore) LoadEdges(ctx context.Context, tenantID string) ([]Edge, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT src, dst FROM edges WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Edge{}
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.Src, &e.Dst); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ClearEdges(ctx context.Context, tenantID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM edges WHERE tenant_id=$1`, tenantID)
	return err
}
