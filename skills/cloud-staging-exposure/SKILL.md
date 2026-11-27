---
name: cloud-staging-exposure
description: "Invoke this skill on authorized assessments when recon reveals cloud storage URLs (S3, GCS, Azure Blob), staging/dev/preview hostnames, CDN origin leaks, exposed metadata/config files (.git, .env, backups, Swagger/OpenAPI, actuator/metrics), or container/CI-CD artifacts."
---

# Cloud, Staging & Configuration Exposure Assessment

Authorized verification workflow for cloud storage, leaked non-production origins, and exposed operational endpoints.

## Context Hygiene (Local-LLM First)
- Save directory listings, bucket XML, and config responses to files in the intent workdir (`cloud_*.txt`) and summarize key counts/entries rather than pasting full XML/JSON listings.

## Verification Playbook

### 1. Staging, Dev & Origin Host Leaks
- When production HTML/JS/headers reference non-production hosts (`staging.*`, `dev.*`, `qa.*`, `*.srv.media`, internal CDN origins):
  - Verify DNS resolution and HTTP reachability within authorized scope.
  - Check for weaker authentication, debug banners, verbose stack traces, exposed admin panels, or stale/unpatched software versions compared to production.
  - Check whether production cookies/sessions are accepted on staging or vice versa (cookie `Domain` scoping).

### 2. Cloud Object Storage (`S3`, `GCS`, `Azure Blob`)
- Check anonymous `GET` listing (`?list-type=2` / container listing) and object readability.
- Verify whether sensitive backups, exports, invoices, or source archives are publicly reachable (read-only checks only; never write/modify objects unless explicitly authorized).

### 3. Operational & Debug Surface
- Check strictly in-scope paths for exposed diagnostics and metadata:
  - OpenAPI/Swagger (`/openapi.json`, `/swagger.json`, `/v3/api-docs`), GraphQL introspection
  - Health/metrics/debug endpoints (`/actuator/*`, `/metrics`, `/debug/pprof`, `/server-status`)
  - VCS/config/backup artifacts (`/.git/HEAD`, `/.env`, `/robots.txt`, `/sitemap.xml`)

### 4. Write-Back & Chaining
- Register newly confirmed in-scope hosts/services via `insert_assets`.
- Record confirmed exposures and negative checks via `record_fact`.
- Report verified exposures with concrete impact (data sensitivity, version/debug leakage, or auth bypass) via `report_finding` with `wstg` / `cwe` IDs from `knowledge_lookup`.
