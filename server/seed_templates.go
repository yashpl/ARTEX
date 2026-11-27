package server

import (
	"errors"
	"log"
	"strings"

	"github.com/Autumn-27/artex/db"
)

type builtinTaskTemplateSeed struct {
	Name        string
	Category    string
	Description string
	Goal        string
}

var builtinTaskCategorySeeds = []string{
	"Web Application",
	"API Security",
	"Reconnaissance",
	"Auth & Session",
	"Cloud & Infrastructure",
	"Code & Binary Audit",
}

var builtinTaskTemplateSeeds = []builtinTaskTemplateSeed{
	{
		Name:        "Full Web Application Assessment",
		Category:    "Web Application",
		Description: "End-to-end authorized web application security assessment covering surface mapping, access control, injection, business logic, and verified PoC reporting.",
		Goal: `Target: <replace with authorized target URL and scope>

Conduct a structured, end-to-end authorized web application penetration test against the target:

1. Surface Mapping & Fingerprinting
   - Enumerate entry points, routes, forms, API endpoints, upload handlers, and client-side assets.
   - Identify technology stack, framework conventions, authentication boundaries, and user roles.
   - Record discovered hosts, endpoints, and parameters systematically so later rounds do not repeat work.

2. High-Impact Vulnerability Testing
   - Access Control & Multi-Tenant Isolation: test IDOR/BOLA, horizontal/vertical privilege escalation, forced browsing, and missing function-level authorization.
   - Injection & Server-Side Flaws: test SQLi, NoSQLi, command injection, SSTI, SSRF, path traversal/LFI, XXE, and unsafe deserialization on real data-bearing parameters.
   - Business Logic & State Integrity: test workflow step skipping, race conditions, price/quantity/quota manipulation, and file-upload validation bypasses.
   - Client-Side Impact: test reflected/stored/DOM XSS, CSRF on state-changing actions, and open redirects chained with auth flows.

3. Execution & Verification Discipline
   - File-first workflow: redirect large scan outputs, HTML pages, and JS bundles to workspace files and inspect targeted slices instead of flooding context.
   - Verify every candidate finding with a clean, reproducible PoC demonstrating concrete impact (actual data access, unauthorized state change, or code/script execution). Do not report unverified scanner hits or speculative issues as confirmed findings.`,
	},
	{
		Name:        "REST & GraphQL API Security Audit",
		Category:    "API Security",
		Description: "Focused OWASP API Top 10 security audit for REST, JSON-RPC, and GraphQL endpoints, emphasizing object-level auth, schema abuse, and mass assignment.",
		Goal: `Target: <replace with authorized API base URL, OpenAPI/Swagger or GraphQL endpoint, and test credentials if provided>

Perform an authorized API security audit across all exposed REST and GraphQL surfaces:

1. Schema & Endpoint Discovery
   - Locate OpenAPI/Swagger specs, GraphQL introspection or field suggestions, versioned prefixes (/v1, /v2, /internal), and undocumented endpoints referenced in client bundles.
   - Save schemas, sample payloads, and endpoint inventories to workspace files and work from structured slices.

2. Authorization & Object Access (BOLA / BFLA / BOPLA)
   - Test object-level authorization by replaying requests across different user/tenant IDs, UUIDs, and nested resource paths.
   - Test function-level authorization on admin, moderation, export, and batch endpoints across privilege tiers.
   - Test mass assignment and excessive data exposure by injecting privileged fields (role, is_admin, plan, owner_id, verified) and inspecting full response objects.

3. API-Specific Attack Vectors
   - GraphQL: test introspection exposure, authorization bypass on nested resolvers/mutations, batching/aliasing abuse, and query depth/cost DoS indicators (non-destructive).
   - REST & Transport: test HTTP method tampering, content-type switching (JSON/XML/form), filter/sort/search injection, SSRF via webhook or URL fields, and rate-limit bypasses on sensitive endpoints.

4. Verification & Reporting
   - Confirm every finding with paired baseline vs. exploit HTTP requests/responses proving unauthorized read/write or business impact. Keep raw traffic and PoC scripts saved in workspace files for clean verification.`,
	},
	{
		Name:        "SPA & JS Bundle Surface Discovery",
		Category:    "Reconnaissance",
		Description: "Deep client-side reconnaissance of Single Page Applications (React/Vue/Angular/Next.js) to extract hidden routes, API contracts, feature flags, and leaked secrets.",
		Goal: `Target: <replace with authorized SPA URL and in-scope domains>

Map the complete client-side and backing API attack surface of the target Single Page Application:

1. Bundle & Source Map Collection
   - Fetch the entry HTML, manifest files, chunk maps, service workers, and all referenced JavaScript bundles into a dedicated workspace directory.
   - Check for exposed source maps (.js.map), build manifests, Next.js/Nuxt data routes, and environment config objects.

2. File-First Static Extraction
   - Run targeted searches against the saved local bundle files (never dump raw minified JS into context) to extract:
     * API base URLs, REST routes, GraphQL queries/mutations, and WebSocket endpoints.
     * Admin, internal, debug, staging, and feature-flagged frontend routes.
     * Hardcoded API keys, third-party tokens, cloud bucket names, OAuth client IDs, and internal hostnames.
     * Client-side authorization checks, role names, and hidden request parameters.

3. Active Surface Validation
   - Probe extracted endpoints and routes against the live target to distinguish active handlers from dead code, noting HTTP status codes, auth requirements, and response shapes.
   - Register verified endpoints, subdomains, and actionable leads cleanly, and validate any leaked credentials or unauthorized endpoints with minimal, safe proof-of-access requests.`,
	},
	{
		Name:        "Breadth-First Target Reconnaissance",
		Category:    "Reconnaissance",
		Description: "Systematic external attack-surface discovery across subdomains, live HTTP services, ports, technologies, and exposed administrative or debug interfaces.",
		Goal: `Target: <replace with authorized root domain(s), IP ranges, or scope list>

Execute a disciplined breadth-first reconnaissance pass to build an accurate asset inventory and prioritize high-value attack surfaces:

1. Asset & Subdomain Enumeration
   - Enumerate subdomains, DNS records (A/AAAA/CNAME/TXT/MX), virtual hosts, and related in-scope hosts using passive and active techniques.
   - Deduplicate and resolve live hosts; save all raw enumeration lists to workspace files before filtering.

2. Service & Web Fingerprinting
   - Probe live web services for status codes, titles, redirects, TLS certificate SANs, server/framework headers, and reverse-proxy behavior.
   - Identify non-standard ports, exposed staging/dev/qa environments, admin consoles, metrics/health endpoints, Swagger/GraphQL docs, and legacy apps.

3. High-Signal Surface Triage
   - Check for low-hanging exposures: directory listings, backup/config files, .git/.env leaks, default landing pages, and dangling CNAMEs.
   - Record all live assets with clear technology tags and prioritize the top attack targets (auth portals, custom APIs, file uploaders, unauthenticated admin panels) with concise evidence.`,
	},
	{
		Name:        "Authentication, OAuth & Session Audit",
		Category:    "Auth & Session",
		Description: "Targeted security review of login, registration, MFA, password reset, JWT/cookie sessions, and OAuth/OIDC/SSO flows.",
		Goal: `Target: <replace with authorized auth/SSO URLs and test account details if available>

Audit all authentication, federation, and session-management controls on the target:

1. Auth & Account Lifecycle Flows
   - Map registration, login, password reset, email verification, invite links, account linking, and MFA enrollment/challenge flows.
   - Test for account enumeration, weak/predictable reset tokens, host-header poisoning in reset links, MFA bypass (step skipping, response manipulation, backup code reuse), and brute-force/credential-stuffing controls.

2. Session & Token Security
   - Inspect cookies (HttpOnly, Secure, SameSite, domain/path scope), session fixation, logout invalidation, and concurrent session handling.
   - For JWTs and opaque tokens: test signature verification, alg=none/key confusion, kid/jku/x5u header injection, weak HMAC secrets, claim tampering, and token reuse after password change or revocation.

3. OAuth 2.0 / OIDC / SSO
   - Test redirect_uri validation bypasses, missing or unvalidated state/nonce/PKCE parameters, authorization code replay, token leakage via Referer/fragment, and pre-account takeover via unverified social/SSO email linking.

4. Verification Discipline
   - Save captured auth requests/tokens to workspace files and prove every vulnerability with a complete, reproducible PoC showing unauthorized account access, session hijacking, or bypass of a security boundary.`,
	},
	{
		Name:        "Cloud, Staging & Exposure Review",
		Category:    "Cloud & Infrastructure",
		Description: "Authorized infrastructure and cloud exposure assessment targeting public storage, SSRF-to-metadata paths, CI/CD artifacts, reverse proxies, and misconfigured services.",
		Goal: `Target: <replace with authorized domains, cloud endpoints, or IP/CIDR scope>

Assess cloud, staging, and infrastructure-layer exposures across the authorized target surface:

1. Cloud & Storage Surface Discovery
   - Identify cloud providers, CDN/WAF edges, origin IPs, object storage buckets (S3/GCS/Azure Blob), container registries, and serverless/API gateway endpoints referenced in DNS, headers, or client assets.
   - Test storage and public artifact endpoints for unauthenticated listing, readable sensitive objects, or overly permissive write ACLs (non-destructive verification only).

2. Infrastructure & Proxy Misconfigurations
   - Check staging/dev/internal subdomains for exposed dashboards, Prometheus/Actuator/pprof/metrics endpoints, CI/CD webhooks, Kubernetes/Docker APIs, or debug panels.
   - Test reverse-proxy and routing boundaries: host-header routing mismatches, path normalization differentials (../, ;, encoded slashes), cache poisoning/deception, and hop-by-hop header abuse.

3. SSRF & Credential Exposure
   - Where URL fetchers, webhooks, PDF/image renderers, or import features exist, test safely for SSRF toward internal services and cloud metadata endpoints (IMDSv1/v2) without modifying cloud state.
   - Save probe outputs to workspace files and report only verified exposures backed by concrete response evidence and clear remediation steps.`,
	},
	{
		Name:        "Client-Side & Browser Security Review",
		Category:    "Web Application",
		Description: "Browser-focused security audit covering DOM/stored/reflected XSS, postMessage abuse, CSRF, clickjacking, CSP bypasses, and sensitive client storage.",
		Goal: `Target: <replace with authorized web application URL and user roles/flows>

Conduct a focused client-side and browser security review of the target application:

1. Input-to-Sink & XSS Analysis
   - Identify all user-controlled sources (URL params, hash fragments, storage, server-reflected fields, rich-text/Markdown editors, file names, postMessage payloads).
   - Trace dangerous DOM sinks (innerHTML, dangerouslySetInnerHTML, document.write, eval, dynamic script/iframe src, template injection) by saving client scripts to workspace files and inspecting targeted matches.
   - Test reflected, stored, and DOM-based XSS and evaluate whether CSP, Trusted Types, or sanitizers can be bypassed cleanly.

2. Cross-Origin & Browser Isolation
   - Audit CORS configurations (reflected Origin with Allow-Credentials, null origin, weak subdomain regexes) and verify whether sensitive authenticated data is readable cross-origin.
   - Inspect window.postMessage listeners for missing origin validation or unsafe message handling.
   - Check state-changing endpoints for CSRF (missing tokens, token reuse, SameSite gaps, GET-based state changes) and clickjacking on sensitive actions.

3. Client Storage & Verification
   - Check localStorage, sessionStorage, IndexedDB, and URL parameters for leaked bearer tokens, PII, or session secrets.
   - Verify every finding in a real browser or reproducible HTTP/HTML PoC showing concrete exploitability rather than reporting missing headers alone.`,
	},
	{
		Name:        "White-Box Source & Dependency Audit",
		Category:    "Code & Binary Audit",
		Description: "Source-code and dependency security audit combining static sink-to-source tracing, auth middleware review, secret scanning, and live PoC verification where applicable.",
		Goal: `Target: <replace with local repository path, archive path, or authorized target URL + source location>

Perform a thorough white-box security audit of the provided source code, configuration, and dependencies:

1. Architecture & Trust-Boundary Mapping
   - Map repository layout, entry points, HTTP/RPC routers, authentication/authorization middleware, database/ORM layer, file handling, and external integrations.
   - Use file-first targeted searches (ripgrep/grep/AST queries) and read only relevant functions and surrounding context; write notes and call-path maps to workspace files.

2. High-Impact Code Audit
   - Authorization & Auth Logic: trace every route/handler to verify consistent authentication, role checks, and tenant/object ownership enforcement.
   - Dangerous Sinks: audit raw SQL/query builders, shell/exec invocations, template rendering, deserialization, file path joins, XML parsers, and outbound HTTP clients (SSRF).
   - Cryptography, Secrets & Config: check for hardcoded secrets, weak randomness, insecure JWT/session settings, debug flags, and permissive CORS/TLS settings.
   - Dependencies & Build: inspect lockfiles/manifests for known high-impact vulnerable packages and verify whether vulnerable code paths are actually reachable.

3. Proof & Reporting
   - For every finding, cite exact file paths, line numbers, data flow from untrusted source to sink, and a concrete PoC (unit test, exploit script, or live HTTP request if a running instance is in scope).`,
	},
}

func normalizeTemplateSeedKey(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

// seedBuiltinTaskTemplates seeds the built-in Task Categories and Task Templates
// ONCE per database (guarded by a settings flag). Existing user-created
// categories or templates with matching names are preserved, and if a user later
// edits or deletes any seeded category or template it is never overwritten or
// recreated on restart.
func (s *Server) seedBuiltinTaskTemplates() {
	if s.m == nil || s.m.pg == nil {
		return
	}
	const flag = "builtin_task_templates_seed_v1"
	if v, _, _ := s.m.pg.GetSetting(flag); v == "true" {
		return
	}

	existingCats, err := s.m.pg.ListTaskCategories()
	if err != nil {
		log.Printf("[templates] failed to list task categories during seed: %v", err)
		return
	}
	catIDByNKey := make(map[string]int64, len(existingCats)+len(builtinTaskCategorySeeds))
	for _, c := range existingCats {
		nkey := c.NKey
		if nkey == "" {
			nkey = normalizeTemplateSeedKey(c.Name)
		}
		catIDByNKey[nkey] = c.ID
	}

	createdCats := 0
	for _, name := range builtinTaskCategorySeeds {
		nkey := normalizeTemplateSeedKey(name)
		if _, exists := catIDByNKey[nkey]; exists {
			continue
		}
		cat, err := s.m.pg.CreateTaskCategory(name)
		if err == nil {
			catIDByNKey[nkey] = cat.ID
			createdCats++
			continue
		}
		if errors.Is(err, db.ErrTaskCategoryNameConflict) {
			if refreshed, listErr := s.m.pg.ListTaskCategories(); listErr == nil {
				for _, c := range refreshed {
					k := c.NKey
					if k == "" {
						k = normalizeTemplateSeedKey(c.Name)
					}
					catIDByNKey[k] = c.ID
				}
			}
			continue
		}
		log.Printf("[templates] failed to seed task category %q: %v", name, err)
		return
	}

	createdTmpls := 0
	for _, seed := range builtinTaskTemplateSeeds {
		var catID *int64
		if id, ok := catIDByNKey[normalizeTemplateSeedKey(seed.Category)]; ok {
			cid := id
			catID = &cid
		}
		_, err := s.m.pg.CreateTaskTemplate(db.TaskTemplateInput{
			Name:        seed.Name,
			Description: seed.Description,
			Goal:        seed.Goal,
			CategoryID:  catID,
		})
		if err == nil {
			createdTmpls++
			continue
		}
		if errors.Is(err, db.ErrTaskTemplateNameConflict) {
			continue
		}
		log.Printf("[templates] failed to seed task template %q: %v", seed.Name, err)
		return
	}

	if err := s.m.pg.SetSetting(flag, "true"); err != nil {
		log.Printf("[templates] failed to record seed flag %s: %v", flag, err)
		return
	}
	if createdCats > 0 || createdTmpls > 0 {
		log.Printf("[templates] seeded %d built-in task categories and %d task templates (one-time)", createdCats, createdTmpls)
	}
}
