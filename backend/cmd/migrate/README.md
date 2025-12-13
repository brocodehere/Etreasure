# Database Migrations

This directory contains the Go-based migration system for the Etreasure backend.

## Overview

The migration system creates all necessary database tables and schemas that match the handler code structure. It uses pgxpool for database connections and supports rolling migrations up and down.

## Available Migrations

1. **001** - Create core authentication and user tables (users, roles, user_roles)
2. **002** - Create categories table
3. **003** - Create products and related tables (products, product_variants, product_categories)
4. **004** - Create media and product images tables (media, product_images)
5. **005** - Create banners table (matching handler schema with UUID primary keys)
6. **006** - Create offers table (matching handler schema)
7. **007** - Create customers and addresses tables (matching handler schema)
8. **008** - Create orders and order items tables (matching handler schema)
9. **009** - Create inventory tables (inventory_items, inventory_adjustments)
10. **010** - Create settings table with default values
11. **011** - Create audit logs table

## Usage

### Using Make commands (recommended)

```bash
# Set your database URL in .env file or environment
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Run all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Check migration status
make migrate-status

# Complete development setup (runs migrations)
make dev-setup

# Reset database (rollback and re-run all migrations)
make reset-db
```

### Using Go commands directly

```bash
# Run migrations
go run ./cmd/migrate up "postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Rollback migrations
go run ./cmd/migrate down "postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Check status
go run ./cmd/migrate status "postgres://user:password@localhost:5432/dbname?sslmode=disable"
```

### Using environment variables

```bash
# Set DATABASE_URL environment variable
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Run commands without specifying URL
go run ./cmd/migrate up
go run ./cmd/migrate status
```

## Database Schema

The migration system creates tables that match the handler expectations:

- **UUID primary keys** for most entities (banners, offers, customers, orders, etc.)
- **SERIAL primary keys** for reference tables (users, roles, products, categories)
- **Proper foreign key relationships** with correct data types
- **Indexes** for performance optimization
- **Default data** for roles and settings

## Features

- **Idempotent migrations** - Uses `IF NOT EXISTS` clauses
- **Rollback support** - Each migration has up and down functions
- **Migration tracking** - Tracks applied migrations in `schema_migrations` table
- **Transaction safety** - Each migration runs in a transaction
- **Error handling** - Proper error reporting and rollback on failure

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (required)

## Example .env file

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable
JWT_SECRET=your-secret-key
REFRESH_SECRET=your-refresh-secret
BACKEND_PORT=8080
```

## Troubleshooting

### Connection Issues
- Ensure PostgreSQL is running
- Check the DATABASE_URL format
- Verify database exists and user has permissions

### Migration Failures
- Check the error message for specific issues
- Use `make migrate-status` to see current state
- Use `make migrate-down` to rollback failed migrations
- Re-run with `make migrate-up`

### Foreign Key Constraints
- The system handles mixed UUID and SERIAL primary keys correctly
- Product and variant IDs are SERIAL (INT) in orders and inventory
- Customer and order IDs are UUID for better scalability

## Development

When adding new migrations:

1. Add a new migration in `migrations.go` using `m.AddMigration()`
2. Use the next sequential version number
3. Implement both `up` and `down` functions
4. Test both directions
5. Update this README

The migration system ensures your database schema stays in sync with your application code.
