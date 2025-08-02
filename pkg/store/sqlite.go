package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bradtumy/authorization-service/pkg/policy"
	"github.com/bradtumy/authorization-service/pkg/tenant"
)

// SQLiteStore implements Store backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite creates a new SQLiteStore using the given datasource name.
func NewSQLite(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) SaveTenant(ctx context.Context, t tenant.Tenant) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO tenants(id, name, created_at) VALUES(?,?,?)`, t.ID, t.Name, t.CreatedAt.Unix())
	return err
}

func (s *SQLiteStore) LoadTenant(ctx context.Context, id string) (tenant.Tenant, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM tenants WHERE id=?`, id)
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

func (s *SQLiteStore) ListTenants(ctx context.Context) ([]tenant.Tenant, error) {
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

func (s *SQLiteStore) DeleteTenant(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tenants WHERE id=?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM policies WHERE tenant_id=?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM edges WHERE tenant_id=?`, id)
	return err
}

func (s *SQLiteStore) SavePolicy(ctx context.Context, tenantID string, p policy.Policy) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT OR REPLACE INTO policies(tenant_id, policy_id, policy) VALUES(?,?,?)`, tenantID, p.ID, string(b))
	return err
}

func (s *SQLiteStore) LoadPolicies(ctx context.Context, tenantID string) ([]policy.Policy, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT policy FROM policies WHERE tenant_id=?`, tenantID)
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

func (s *SQLiteStore) ClearPolicies(ctx context.Context, tenantID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM policies WHERE tenant_id=?`, tenantID)
	return err
}

func (s *SQLiteStore) SaveEdge(ctx context.Context, tenantID, src, dst string) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO edges(tenant_id, src, dst) VALUES(?,?,?)`, tenantID, src, dst)
	return err
}

func (s *SQLiteStore) LoadEdges(ctx context.Context, tenantID string) ([]Edge, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT src, dst FROM edges WHERE tenant_id=?`, tenantID)
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

func (s *SQLiteStore) ClearEdges(ctx context.Context, tenantID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM edges WHERE tenant_id=?`, tenantID)
	return err
}
