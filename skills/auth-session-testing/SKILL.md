---
name: auth-session-testing
description: "Invoke this skill during authorized web or API assessments to evaluate authentication, session management, OAuth/OIDC, JWT handling, password recovery, MFA flows, and cross-account authorization (IDOR/BOLA/BFLA). Optimized for concise, file-first execution on local LLMs."
---

# Authentication & Session Security Testing

Authorized testing methodology for authentication, session lifecycle, and authorization boundaries (mapped to OWASP WSTG-ATHN / WSTG-SESS / WSTG-ATHZ).

## Context Hygiene (Local-LLM First)
- Save raw HTTP responses (`>4 KB`) to files in the current intent workdir (`auth_*.txt` / `auth_*.json`) and inspect only status lines, headers, and diff summaries with `grep`/`jq`.
- Never dump full HTML pages or large JWT/cookie lists into conversation text.

## Systematic Test Matrix

### 1. Session & Cookie Controls (`WSTG-SESS-02`, `WSTG-SESS-03`)
- Inspect `Set-Cookie` attributes: `HttpOnly`, `Secure`, `SameSite`, `Domain`, `Path`, `Max-Age`.
- Test **session rotation** before vs. after login and privilege change.
- Test **logout invalidation**: replay the pre-logout cookie/token against a protected endpoint and confirm whether the server rejects it.

### 2. Token & JWT Validation (`WSTG-SESS-10`, `CWE-347`)
- Decode header/payload locally (`jq`/Python) and record `alg`, `kid`, `iss`, `aud`, `exp`, `sub`, and role claims.
- Verify signature enforcement: tampered payload, mismatched algorithm, missing signature, expired token (`exp`), and wrong audience/issuer.
- Check whether `kid` or `jku`/`x5u` headers are accepted or validated strictly against an allowlist.

### 3. OAuth 2.0 / OIDC Flows (`WSTG-ATHN-05`, `CWE-601`)
- Verify `redirect_uri` exact-match enforcement (path traversal, subdomain, parameter pollution).
- Check `state` / `nonce` / PKCE (`code_verifier`) binding and single-use enforcement on callbacks.
- Check whether authorization codes or tokens leak via `Referer` headers or open redirects.

### 4. Account Recovery & MFA (`WSTG-ATHN-09`, `WSTG-ATHN-10`)
- Check password-reset endpoints for `Host` / `X-Forwarded-Host` header trust, token entropy/expiry, and account-enumeration response differentials.
- Verify MFA step enforcement: test whether protected API endpoints accept a pre-MFA session token.

### 5. Cross-Account Authorization Matrix (`WSTG-ATHZ-04`, `CWE-639`)
- For every object-scoped endpoint (`/orders/{id}`, `/users/{id}`, `/invoices/{id}`), compare responses across:
  1. Unauthenticated request
  2. Authenticated User A accessing User A's resource (baseline)
  3. Authenticated User A accessing User B's resource (horizontal)
  4. Low-privilege user accessing admin/staff actions (vertical)

## Write-Back Checklist
- Record each verified boundary check (both confirmed flaws and negative controls) via `record_fact` with exact HTTP method, path, status code, and minimal diff evidence.
- Report confirmed vulnerabilities via `report_finding` citing the matching `knowledge_lookup` `wstg` / `cwe` / `owasp-api` IDs.
