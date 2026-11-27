---
name: js-bundle-mining
description: "Invoke this skill on authorized web assessments to extract hidden API routes, GraphQL operations, feature flags, admin paths, sourcemaps, and exposed client-side configuration or credentials from JavaScript bundles (Webpack, Vite, Next.js, Nuxt, Angular). Uses file-first scripted extraction to keep local-LLM contexts compact."
---

# JavaScript Bundle & Frontend Asset Mining

Fast, script-driven extraction of attack surface from frontend JavaScript bundles without bloating the LLM context window.

## Hard Context Rules (Local-LLM First)
- **Never** read minified `.js` bundles (`>50 KB`) directly into context.
- Download bundles into `js/` inside the current intent workdir and run concise Python/Bash extractors that write deduplicated summaries to `js_endpoints.txt`, `js_secrets.txt`, and `js_maps.txt`.
- Only inspect small (`<=30` line) context windows around confirmed hits.

## Workflow

### 1. Collect Bundle & Manifest Inventory
- Fetch entry HTML and extract `<script src>`, `<link rel="modulepreload">`, Next.js `_buildManifest.js` / `_ssgManifest.js`, and Vite/Webpack chunk maps.
- Check for publicly accessible `.js.map` source maps (`curl -sI <bundle>.map` → `200 OK`).

### 2. Scripted Endpoint & Route Extraction
- Extract candidate paths (`/api/...`, `/v1/...`, `/graphql`, `/admin/...`, internal service hostnames, WebSocket `ws://`/`wss://` URLs) into `js_endpoints.txt` and deduplicate (`sort -u`).
- Extract GraphQL query/mutation/fragment names and variable schemas into `js_graphql.txt`.
- Extract environment/config objects (`NEXT_PUBLIC_*`, `VITE_*`, `REACT_APP_*`, `window.__INITIAL_STATE__`, `__NEXT_DATA__`).

### 3. Client-Side Secret & Staging Host Audit
- Scan bundles for staging/dev/internal hostnames, storage bucket URLs (`s3.amazonaws.com`, `storage.googleapis.com`, `blob.core.windows.net`), and high-entropy tokens/keys.
- Distinguish intentionally public client keys (e.g. analytics/publishable keys) from privileged or misconfigured secrets by checking scope/permissions safely.

### 4. Write-Back
- Batch-insert verified endpoints and discovered hosts with `insert_assets`.
- Record compact summary facts (`record_fact`) citing file paths and counts (e.g. `14 new /api/v2 routes extracted from chunk-a1b2.js`).
- Report exposed source maps containing sensitive internal logic, staging origins, or privileged keys via `report_finding`.
