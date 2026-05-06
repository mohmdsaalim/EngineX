-- Restore DEFAULT for id
ALTER TABLE orders ALTER COLUMN id SET DEFAULT uuid_generate_v4();