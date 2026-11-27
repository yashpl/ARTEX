---
name: client-side-browser
description: "Invoke this skill on authorized web assessments to test browser-side vulnerabilities using playwright-cli: DOM/reflected/stored XSS, postMessage origin validation flaws, CORS misconfigurations with credentials, clickjacking, CSP weaknesses, and client-side redirect chains."
---

# Client-Side & Browser Security Testing

Authorized browser-verified testing workflow pairing `playwright-cli` with strict DOM/console/network proof (mapped to OWASP WSTG-CLNT).

## Context Hygiene (Local-LLM First)
- Use `playwright-cli --raw eval ...` and targeted `playwright-cli find` instead of capturing full-page snapshots on large sites.
- Save DOM/script dumps to files in the intent workdir and summarize findings concisely.

## Test Areas

### 1. Reflection & DOM Sinks (`WSTG-CLNT-01`, `WSTG-CLNT-02`, `CWE-79`)
- Map untrusted sources (`location.search`, `location.hash`, `document.referrer`, `window.name`, `postMessage`) to dangerous sinks (`innerHTML`, `outerHTML`, `document.write`, `eval`, `setTimeout`, `location.href`).
- Probe with harmless canary markers first to determine exact HTML/attribute/JS-string context and encoding behavior before testing minimal proof payloads.
- Confirm execution directly in headless Chromium via `playwright-cli console` or a deterministic DOM marker (never assume reflection equals execution).

### 2. Cross-Origin & Messaging Controls (`WSTG-CLNT-07`, `WSTG-CLNT-11`)
- **CORS:** Test `Origin: https://attacker.example` and subdomain/null variants against authenticated endpoints; verify whether `Access-Control-Allow-Origin` reflects the untrusted origin together with `Access-Control-Allow-Credentials: true`.
- **`postMessage`:** Audit `window.addEventListener('message', ...)` handlers in bundled JS for missing or weak `event.origin` checks.

### 3. Framing, CSP & Client Redirects (`WSTG-CLNT-04`, `WSTG-CLNT-09`)
- Inspect `Content-Security-Policy` (`unsafe-inline`, `unsafe-eval`, wildcard `script-src`, missing `base-uri`/`frame-ancestors`) and `X-Frame-Options`.
- Test redirect parameters (`next`, `returnUrl`, `redirect`, `continue`) for open redirects and `javascript:`/`data:` scheme handling.

## Write-Back
- Record verified sinks, encodings, and negative checks via `record_fact`.
- Report only browser-confirmed vulnerabilities with reproducible steps and exact HTTP/DOM evidence via `report_finding`.
