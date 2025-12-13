# Search subsystem (Postgres + pg_trgm + tsvector)

This folder contains the search migration and helper code.

How to run migration:

1. Ensure your DB user has permission to create extensions. If using a managed DB, enable `pg_trgm` and `unaccent` in the provider console.
2. From the `backend` directory run your migration tool (or psql):

```powershell
pq: # example using psql
psql "$DATABASE_URL" -f migrations/0001_search_setup.sql
```

Manually rebuild search_vector for all products if needed:

```sql
UPDATE products SET search_vector = (
  setweight(to_tsvector('simple', coalesce(unaccent(title),'')), 'A') ||
  setweight(to_tsvector('simple', coalesce(unaccent(array_to_string(tags, ' '),''))), 'B') ||
  setweight(to_tsvector('simple', coalesce(unaccent(brand),'')), 'B') ||
  setweight(to_tsvector('simple', coalesce(unaccent(description),'')), 'C')
);
```

If `pg_trgm` is unavailable the handler will still attempt searches but may fall back to less efficient ILIKE searches (not implemented automatically).

Running tests (integration):

```powershell
SET TEST_DATABASE_URL="postgres://user:pass@localhost:5432/etest?sslmode=disable"
go test ./internal/search -v
go test ./cmd/api/handlers -v
```
# Etreasure Admin Backend

Go-based backend for the Etreasure admin panel.

## Quick start

1. Copy `.env.sample` to `.env` and adjust values.
2. Run PostgreSQL locally and create database `etreasure`.
3. Apply migrations (e.g. using golang-migrate or your preferred tool) from the `migrations/` directory.
4. Seed initial data (roles, admin user, sample catalog):

   ```bash
   go run ./cmd/seed
   ```

5. Run the API:

```bash
cd backend
go run ./cmd/api
```

This is an initial scaffold; further endpoints, migrations, and docs will be added.
