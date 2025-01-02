INSERT INTO tenants (tenant_name, contact_email, plan) VALUES ('Acme', 'admin@acme.com', 'free');
INSERT INTO users (user_name, email, tenant_id) VALUES ('admin', 'admin@acme.com', 1);
