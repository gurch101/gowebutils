CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BLOB NOT NULL,
	expiry REAL NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions(expiry);

CREATE TABLE IF NOT EXISTS tenants (
    id INTEGER PRIMARY KEY,         -- Unique ID for each tenant
    tenant_name TEXT NOT NULL CONSTRAINT unique_tenant_name UNIQUE CHECK(tenant_name <> ''),    -- Name of the tenant
    contact_email TEXT NOT NULL CHECK(contact_email <> ''),  -- Contact email of the tenant
    plan TEXT NOT NULL CHECK(plan <> ''),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,        -- Status of the tenant (active/inactive)
    role_id INTEGER REFERENCES roles(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,   -- Timestamp when the tenant was created
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,   -- Timestamp when the tenant was last updated
    version INTEGER NOT NULL DEFAULT 1 CHECK (version >= 1)
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,           -- Unique ID for each user
    user_name VARCHAR(255) NOT NULL,           -- User's name
    email TEXT NOT NULL UNIQUE CHECK(email <> ''),   -- Unique email for each user
    tenant_id INTEGER NOT NULL,           -- Foreign key to tenants table
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,   -- Timestamp when the tenant was created
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,   -- Timestamp when the user was last updated
    CONSTRAINT fk_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants (id)    -- Foreign key referencing tenants table
        ON DELETE CASCADE                 -- Optional: Deletes the user if tenant is deleted
);

-- -- Create an index on tenant_id and name
CREATE INDEX IF NOT EXISTS idx_tenant_id_name ON users (tenant_id, user_name);

-- -- Create an index on tenant_id and email
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_id_email_unique ON users (tenant_id, email);

CREATE TABLE IF NOT EXISTS user_login_attempts (
    id BIGINT PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS user_login_attempts_user_id_idx ON user_login_attempts (user_id, created_at);

CREATE TABLE roles (
    tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE INDEX IF NOT EXISTS roles_idx ON roles (tenant_id, role_name);

CREATE TABLE role_permissions (
    tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    code VARCHAR(255) NOT NULL,  -- e.g., "create_post", "delete_user"
    description TEXT,
    PRIMARY KEY (role_id, code)
);

CREATE INDEX IF NOT EXISTS role_permissions_idx ON role_permissions (tenant_id, role_id, code);
