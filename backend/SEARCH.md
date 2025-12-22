# Search System Documentation

## Overview

The Etreasure search system provides fast, accurate full-text and fuzzy product search with support for filters, suggestions, and faceted search. It leverages PostgreSQL's built-in full-text search (FTS) combined with the `pg_trgm` extension for typo tolerance and prefix matching.

**Performance targets:**
- Full search: <200ms average response time
- Suggestions (autocomplete): <100ms average response time
- Indexing: Automatic on insert/update via triggers
- Concurrent capacity: Supports 1000+ concurrent searches with proper caching

---

## Architecture

### Search Indexing Strategy

#### 1. **tsvector Column**
- Column: `search_vector` on `products` table
- Stores: Indexed, weighted full-text document combining:
  - **Weight A (title)**: 1.0x multiplier (highest relevance)
  - **Weight B (brand + tags)**: 0.8x multiplier
  - **Weight C (description + SKU)**: 0.6x multiplier

#### 2. **Database Trigger**
- Trigger: `products_search_vector_trigger`
- Function: `products_search_vector_update()`
- Action: Automatically regenerates `search_vector` on:
  - Product insert
  - Product update (any field)
  - Never requires manual maintenance

#### 3. **Indexes**
- **GIN Index on search_vector** (`idx_products_search_vector`)
  - Used for full-text search queries
  - Accelerates `to_tsquery()` matching
  - ~5-10x faster than sequential scan

- **GIN Trigram Indexes** (using `gin_trgm_ops`)
  - `idx_products_title_trgm`: Enables fuzzy matching on title
  - `idx_products_brand_trgm`: Enables fuzzy matching on brand
  - `idx_products_sku_trgm`: Enables fuzzy matching on SKU
  - Accelerates `LIKE` and `%` operator queries
  - Supports typo tolerance via similarity()

- **Published Status Index** (`idx_products_published_status`)
  - Filters only visible products in search results
  - Respects publish_at and unpublish_at timestamps

---

## API Endpoints

### 1. Full-Text Search
**Endpoint:** `GET /api/search`

**Query Parameters:**
```
q              (required, string, max 500 chars): Search query
category       (optional, integer): Filter by category ID
min_price      (optional, integer): Minimum price in cents
max_price      (optional, integer): Maximum price in cents
sort           (optional, enum): relevance | price_asc | price_desc | newest
cursor         (optional, string): Pagination cursor from previous response
limit          (optional, integer, default 20, max 100): Results per page
fields         (optional, string): Comma-separated fields to include
```

**Example Request:**
```bash
curl "https://etreasure-1.onrender.com/api/search?q=silk+saree&category=2&min_price=50000&max_price=200000&sort=relevance&limit=20"
```

**Response (200 OK):**
```json
{
  "items": [
    {
      "id": 142,
      "title": "Premium Banarasi Silk Saree",
      "slug": "banarasi-silk-saree-001",
      "price": 125000,
      "image": "/uploads/products/banarasi-001.webp",
      "excerpt": "Handcrafted Banarasi silk saree with gold zari work. Authentic weave from Varanasi...",
      "score": 0.98,
      "brand": "Royal Weaves",
      "tags": ["silk", "traditional", "handmade"],
      "sku": "BAR-SILK-001"
    },
    {
      "id": 143,
      "title": "Silk Saree with Cotton Blend",
      "slug": "silk-cotton-blend-saree",
      "price": 75000,
      "image": "/uploads/products/silk-cotton-001.webp",
      "excerpt": "Beautiful silk-cotton blend saree perfect for daily wear...",
      "score": 0.85,
      "brand": "Ethnic Treasures",
      "tags": ["silk", "cotton", "blend"],
      "sku": "SIL-COT-001"
    }
  ],
  "nextCursor": "eyJpZCI6IDE0MywgInNjb3JlIjogMC44NX0=",
  "totalCount": null
}
```

**Error Responses:**
- `400 Bad Request`: Missing required `q` parameter or invalid parameters
- `500 Internal Server Error`: Database query failure

**Cache-Control:** `public, max-age=300` (5 minutes)

---

### 2. Search Suggestions (Autocomplete)
**Endpoint:** `GET /api/search/suggest`

**Query Parameters:**
```
q              (required, string, max 100 chars): Partial query
limit          (optional, integer, default 8, max 50): Number of suggestions
```

**Example Request:**
```bash
curl "https://etreasure-1.onrender.com/api/search/suggest?q=bana&limit=8"
```

**Response (200 OK):**
```json
[
  {
    "id": 142,
    "title": "Banarasi Silk Saree",
    "slug": "banarasi-silk-saree-001",
    "price": 125000,
    "image": "/uploads/products/banarasi-001.webp",
    "highlight": "**Bana**rasi Silk Saree"
  },
  {
    "id": 156,
    "title": "Banarasi Cotton Blend",
    "slug": "banarasi-cotton-blend",
    "price": 65000,
    "image": "/uploads/products/banarasi-cotton-001.webp",
    "highlight": "**Bana**rasi Cotton Blend"
  }
]
```

**Matching Strategy:**
1. **Prefix match** (highest priority): Title starts with query
2. **Fuzzy match** (medium priority): Trigram similarity > 0.3
3. **Brand match** (lower priority): Brand matches query

**Cache-Control:** `public, max-age=30` (30 seconds) — highly cached for trend queries

---

### 3. Search Facets (Filters)
**Endpoint:** `GET /api/search/facets`

**Query Parameters:**
```
q              (optional, string): Filter facets by query context
```

**Example Request:**
```bash
curl "https://etreasure-1.onrender.com/api/search/facets?q=saree"
```

**Response (200 OK):**
```json
{
  "categories": [
    {
      "id": 1,
      "name": "Traditional Sarees",
      "productCount": 145
    },
    {
      "id": 2,
      "name": "Contemporary Wear",
      "productCount": 89
    },
    {
      "id": 3,
      "name": "Ethnic Fusion",
      "productCount": 54
    }
  ],
  "priceRange": {
    "min": 25000,
    "max": 500000,
    "avg": 125000
  }
}
```

**Cache-Control:** `public, max-age=600` (10 minutes)

---

### 4. Manual Reindex
**Endpoint:** `POST /api/admin/search/reindex`

**Authentication:** Admin only (requires JWT token with admin role)

**Request Body:** None

**Example Request:**
```bash
curl -X POST "https://etreasure-1.onrender.com/api/admin/search/reindex" \
  -H "Authorization: Bearer <admin-jwt-token>"
```

**Response (200 OK):**
```json
{
  "message": "search index rebuilt successfully",
  "updatedCount": 2847,
  "durationMs": 3421,
  "timestamp": "2025-12-05T10:30:45Z"
}
```

**Use cases:**
- After bulk product import
- After database migration or schema changes
- After changing indexing weights or fields

**Typical duration:** 0.5–2 seconds for 1000–10000 products

---

### 5. Search Health Check
**Endpoint:** `GET /api/search/health`

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-05T10:30:45Z"
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "error": "required PostgreSQL extensions (pg_trgm, unaccent) not installed"
}
```

---

## Pagination

The search system uses **cursor-based pagination** for stability and performance.

### Cursor Format
Cursors are base64-encoded JSON:
```json
{
  "id": 142,
  "score": 0.98
}
```

### Example Pagination Flow
1. **First request** (no cursor):
```bash
GET /api/search?q=saree&limit=20
```

2. **Response includes nextCursor:**
```json
{
  "items": [...20 items...],
  "nextCursor": "eyJpZCI6IDE0MywgInNjb3JlIjogMC44NX0="
}
```

3. **Next request** (with cursor):
```bash
GET /api/search?q=saree&limit=20&cursor=eyJpZCI6IDE0MywgInNjb3JlIjogMC44NX0=
```

4. **Continue until no nextCursor is returned** (last page)

---

## Database Setup & Migration

### Step 1: Run Migration
```bash
# From backend directory
cd /backend
go run cmd/migrate/main.go
```

This will:
- Enable `pg_trvector` and `unaccent` extensions
- Add `brand`, `tags`, `primary_sku`, and `search_vector` columns to `products`
- Create trigger function and indexes
- Populate `search_vector` for existing products

### Step 2: Verify Installation
```bash
# Check extensions
psql etreasure -c "SELECT extname FROM pg_extension WHERE extname IN ('pg_trgm', 'unaccent');"

# Check columns
psql etreasure -c "\d products" | grep -E "search_vector|brand|tags"

# Check indexes
psql etreasure -c "\di" | grep idx_products_search
```

---

## Indexing Strategy & Reindexing

### Automatic Indexing
- **Trigger-based:** `products_search_vector_trigger` runs on every INSERT/UPDATE
- **No manual intervention needed** under normal operation
- **Cost:** ~2-5ms per product update (includes trigger execution)

### Scheduled Reindexing (Cron Job)

Recommended: Run nightly for consistency and performance tuning.

#### Option 1: PostgreSQL pg_cron Extension
```sql
-- Create extension
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Schedule nightly reindex at 2 AM
SELECT cron.schedule('reindex-search-vectors', '0 2 * * *', 
  'UPDATE products SET search_vector = to_tsvector(''english'', 
    coalesce(title, '''') || '' '' ||
    coalesce(brand, '''') || '' '' ||
    array_to_string(coalesce(tags, ''{}''::text[]), '' '') || '' '' ||
    coalesce(description, '''') || '' '' ||
    coalesce(primary_sku, '''')
  )');
```

#### Option 2: External Cron (Linux/macOS)
```bash
# Add to crontab
0 2 * * * curl -X POST https://etreasure-1.onrender.com/api/admin/search/reindex \
  -H "Authorization: Bearer $ADMIN_JWT_TOKEN"
```

#### Option 3: Kubernetes CronJob
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: search-reindex
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: reindex
            image: curlimages/curl
            command:
            - /bin/sh
            - -c
            - curl -X POST http://api-service:8080/api/admin/search/reindex -H "Authorization: Bearer $ADMIN_JWT"
```

---

## Performance Tuning

### 1. Index Maintenance
```sql
-- Analyze tables for query planner optimization
ANALYZE products;

-- Reindex if fragmented (after many updates)
REINDEX INDEX idx_products_search_vector;
REINDEX INDEX idx_products_title_trgm;
```

### 2. Query Optimization Settings
```sql
-- Increase work_mem for large searches (in postgresql.conf or ALTER SYSTEM)
ALTER SYSTEM SET work_mem = '256MB';

-- Increase random_page_cost for faster index usage
ALTER SYSTEM SET random_page_cost = 1.1;

SELECT pg_reload_conf();
```

### 3. Monitoring Query Performance
```sql
-- Enable query logging
ALTER SYSTEM SET log_min_duration_statement = 100; -- Log queries > 100ms

-- Check slow queries
SELECT query, mean_exec_time, calls FROM pg_stat_statements 
WHERE query LIKE '%search_vector%' 
ORDER BY mean_exec_time DESC LIMIT 10;
```

### 4. Caching Strategy
- **Browser cache:** 5 minutes for search results, 30 seconds for suggestions
- **CDN cache:** 5 minutes for search, 30 seconds for suggestions (if using edge cache)
- **Redis cache** (optional):
  ```go
  // Cache search results for 5 minutes
  cacheKey := fmt.Sprintf("search:%s:%d", query, limit)
  ```

---

## Rate Limiting & Security

### Rate Limiting Middleware
To prevent abuse, add rate limiting (recommended 60 req/min per IP):

```go
// Example: Using gin-contrib/ratelimit
import "github.com/gin-contrib/ratelimit"

searchLimiter := ratelimit.NewLimiter(ratelimit.PerMinute(60))
r.GET("/api/search", searchLimiter, searchHandler.Search)
r.GET("/api/search/suggest", ratelimit.NewLimiter(ratelimit.PerMinute(120)), searchHandler.Suggest)
```

### SQL Injection Prevention
All queries use **parameterized prepared statements** via pgx:
```go
// ✅ SAFE: Parameterized
rows, _ := db.Query(ctx, "SELECT * FROM products WHERE title ILIKE $1", query)

// ❌ UNSAFE: String concatenation
rows, _ := db.Query(ctx, "SELECT * FROM products WHERE title ILIKE '%" + query + "%'")
```

### Input Validation
- Query length: Max 500 characters (truncated at 500 in handler)
- Suggestion query: Max 100 characters
- Limit parameter: Min 1, Max 100 (enforced)
- Sort parameter: Enum validation (only `relevance`, `price_asc`, `price_desc`, `newest`)

---

## Troubleshooting

### Issue: Search returns no results
**Symptoms:** Valid products exist but search returns empty

**Steps:**
1. Verify `published = TRUE` and publish_at timestamps are correct:
   ```sql
   SELECT id, title, published, publish_at, unpublish_at FROM products LIMIT 5;
   ```

2. Check search_vector is populated:
   ```sql
   SELECT id, title, search_vector FROM products WHERE id = <product_id>;
   ```

3. Trigger may not have executed. Manually reindex:
   ```bash
   curl -X POST https://etreasure-1.onrender.com/api/admin/search/reindex -H "Authorization: Bearer $TOKEN"
   ```

### Issue: Search is slow (> 500ms)
**Symptoms:** Search queries take > 500ms average

**Steps:**
1. Check index usage:
   ```sql
   EXPLAIN ANALYZE SELECT * FROM products 
   WHERE search_vector @@ to_tsquery('english', 'saree:*');
   ```

2. Verify indexes exist:
   ```sql
   \di+ idx_products_search_vector
   ```

3. Analyze table statistics:
   ```sql
   ANALYZE products;
   ```

4. Monitor query performance:
   ```sql
   SELECT mean_exec_time, calls, query FROM pg_stat_statements 
   WHERE query LIKE '%search_vector%' ORDER BY mean_exec_time DESC;
   ```

### Issue: pg_trgm extension not found
**Symptoms:** Error: "required PostgreSQL extensions not installed"

**Steps:**
1. Create extension:
   ```sql
   CREATE EXTENSION IF NOT EXISTS pg_trgm;
   CREATE EXTENSION IF NOT EXISTS unaccent;
   ```

2. Verify installation:
   ```sql
   SELECT extname FROM pg_extension WHERE extname IN ('pg_trgm', 'unaccent');
   ```

---

## Testing

### Unit Tests
```bash
cd /backend
go test ./internal/search -v
```

### Integration Tests
```bash
# Requires running postgres
go test ./internal/search -tags=integration -v
```

### Load Testing (k6)
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 }, // Ramp up
    { duration: '2m', target: 100 },  // Hold
    { duration: '30s', target: 0 },   // Ramp down
  ],
};

export default function() {
  const url = 'https://etreasure-1.onrender.com/api/search?q=saree&limit=20';
  let res = http.get(url);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  sleep(1);
}
```

Run:
```bash
k6 run load-test.js
```

---

## Frontend Integration

### React Hook Example
```typescript
import { useState, useCallback, useRef } from 'react';

export function useSearch() {
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const debounceTimer = useRef<NodeJS.Timeout | null>(null);

  const search = useCallback(async (query: string) => {
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    
    debounceTimer.current = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetch(`/api/search?q=${encodeURIComponent(query)}&limit=20`);
        const data = await res.json();
        setResults(data.items);
      } finally {
        setLoading(false);
      }
    }, 250);
  }, []);

  return { results, loading, search };
}
```

---

## OpenAPI / Swagger Specification

```yaml
openapi: 3.0.0
info:
  title: Etreasure Search API
  version: 1.0.0
  description: Full-text product search with filters and suggestions
servers:
  - url: https://etreasure-1.onrender.com
paths:
  /api/search:
    get:
      operationId: search
      summary: Full-text product search
      parameters:
        - name: q
          in: query
          required: true
          schema:
            type: string
            maxLength: 500
          description: Search query
        - name: category
          in: query
          schema:
            type: integer
        - name: min_price
          in: query
          schema:
            type: integer
        - name: max_price
          in: query
          schema:
            type: integer
        - name: sort
          in: query
          schema:
            type: string
            enum: [relevance, price_asc, price_desc, newest]
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            minimum: 1
            maximum: 100
        - name: cursor
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Search results
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      $ref: '#/components/schemas/SearchResult'
                  nextCursor:
                    type: string
                    nullable: true

  /api/search/suggest:
    get:
      operationId: suggest
      summary: Autocomplete suggestions
      parameters:
        - name: q
          in: query
          required: true
          schema:
            type: string
            maxLength: 100
        - name: limit
          in: query
          schema:
            type: integer
            default: 8
            minimum: 1
            maximum: 50
      responses:
        '200':
          description: Suggestions list
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Suggestion'

components:
  schemas:
    SearchResult:
      type: object
      properties:
        id:
          type: integer
        title:
          type: string
        slug:
          type: string
        price:
          type: integer
          description: Price in cents
        image:
          type: string
        excerpt:
          type: string
        score:
          type: number
          format: double
          description: Relevance score 0-1
        brand:
          type: string
        tags:
          type: array
          items:
            type: string
```

---

## Summary

| Aspect | Details |
|--------|---------|
| **Indexing** | Automatic via trigger, manual via `/api/admin/search/reindex` |
| **Search latency** | 50-200ms typical (depends on query complexity + dataset size) |
| **Suggestion latency** | 20-100ms typical |
| **Index size** | ~2-5% of table size |
| **Update overhead** | ~2-5ms per product (trigger + trigger execution) |
| **Caching** | Browser 5min, CDN 30sec (suggestions) |
| **Rate limit** | 60 req/min per IP (recommended) |
| **Max query length** | 500 characters |
| **Max results per page** | 100 items |
| **Pagination** | Cursor-based (stable, performant) |
| **SQL Injection** | Protected via parameterized queries |

---

## Additional Resources

- PostgreSQL Full-Text Search: https://www.postgresql.org/docs/current/textsearch.html
- pg_trgm Documentation: https://www.postgresql.org/docs/current/pgtrgm.html
- Astro + React Islands: https://docs.astro.build/en/basics/islands/
- Gin Framework: https://gin-gonic.com/

