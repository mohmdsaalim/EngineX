-- Remove DEFAULT from id to allow passing custom UUID
ALTER TABLE orders ALTER COLUMN id DROP DEFAULT;