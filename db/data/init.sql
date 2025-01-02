INSERT INTO tenants (tenant_name, contact_email, plan) VALUES ('Acme', 'admin@acme.com', 'free');
INSERT INTO tenants (tenant_name, contact_email, plan) VALUES ('Flancrest Enterprises', 'admin@flancrest.com', 'paid');
INSERT INTO users (user_name, email, tenant_id) VALUES ('admin', 'admin@acme.com', 1);
INSERT INTO users (user_name, email, tenant_id) VALUES ('john', 'john@acme.com', 1);
