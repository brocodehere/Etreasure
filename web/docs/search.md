# Frontend Search Contract

This document describes the frontend contract with the backend search API.

Endpoint: `GET /api/search?q=...&limit=12`

Response JSON:

```
{
  "q": "saree",
  "results": [
    { "type": "product", "id": 123, "title": "...", "slug":"...", "image":"...", "price": 1299, "link":"/product/...", "excerpt":"..." }
  ],
  "source": "products",
  "took_ms": 25
}
```

- `type` is one of `product | category | offer | banner` and the frontend should use `link` to route to the appropriate page.
- Minimum search input is 4 characters and must include at least one alphabetic character.

SearchBar island located at `web/src/components/SearchBar.tsx`:
- Debounces input by 250ms
- Only calls API if input satisfies min length + alpha
- Renders dropdown suggestions; first Enter navigates to first result if highlighted, otherwise to `/search?q=...`
