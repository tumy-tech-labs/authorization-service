CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    name TEXT,
    created_at BIGINT
);

CREATE TABLE IF NOT EXISTS policies (
    tenant_id TEXT,
    policy_id TEXT,
    policy TEXT,
    PRIMARY KEY (tenant_id, policy_id)
);

CREATE TABLE IF NOT EXISTS edges (
    tenant_id TEXT,
    src TEXT,
    dst TEXT,
    PRIMARY KEY (tenant_id, src, dst)
);
