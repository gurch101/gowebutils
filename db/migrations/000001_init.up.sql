CREATE TABLE IF NOT EXISTS tenants (
    id INTEGER PRIMARY KEY,         -- Unique ID for each tenant
    tenant_name TEXT NOT NULL CONSTRAINT unique_tenant_name UNIQUE CHECK(tenant_name <> ''),    -- Name of the tenant
    contact_email TEXT NOT NULL CHECK(contact_email <> ''),  -- Contact email of the tenant
    plan TEXT NOT NULL CHECK(plan <> ''),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,        -- Status of the tenant (active/inactive)
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

CREATE TABLE IF NOT EXISTS groups (
    id BIGINT PRIMARY KEY,               -- Unique ID for each group
    group_name TEXT NOT NULL,        -- Name of the group (not null)
    descr TEXT,                    -- Optional description of the group
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- Creation timestamp
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- Last update timestamp
    tenant_id INTEGER NOT NULL,          -- Foreign key to the tenants table
    CONSTRAINT fk_tenant
        FOREIGN KEY (tenant_id)
        REFERENCES tenants (id)
        ON DELETE CASCADE
);

-- -- Create a unique index on tenant_id and name
CREATE UNIQUE INDEX IF NOT EXISTS groups_tenant_id_name_idx ON groups (tenant_id, group_name);

CREATE TABLE IF NOT EXISTS user_groups (
    id BIGINT PRIMARY KEY,               -- Unique ID for each record in user_groups
    user_id BIGINT NOT NULL,            -- Foreign key to users table
    group_id BIGINT NOT NULL,           -- Foreign key to groups table
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- Creation timestamp
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- Last update timestamp
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,               -- Cascade delete if the user is deleted
    CONSTRAINT fk_group
        FOREIGN KEY (group_id)
        REFERENCES groups (id)
        ON DELETE CASCADE,               -- Cascade delete if the group is deleted
    CONSTRAINT user_groups_user_id_group_id_unique
        UNIQUE (user_id, group_id)       -- Ensure the user_id, group_id combination is unique
);
