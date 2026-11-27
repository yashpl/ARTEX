# Local-LLM Optimization Guide (`yashpl/ARTEX`)

This fork of ARTEX is specifically tuned to run at high throughput and high stability on **local self-hosted LLMs** (`llama-server` / `llama.cpp` and Ollama) — tested on a single **NVIDIA RTX 4090 (24 GB VRAM)** serving `Qwen3.8-27B` (`UD-Q4_K_XL`, `qwen35` hybrid architecture).

## Key Local-LLM Enhancements in `yashpl/ARTEX`

1. **GBNF-Safe Tool Schemas (`agent/fact_input.go`):**
   - `record_fact` uses anchored, GBNF-safe regex patterns (`\t\n\x0b\x0c\r` escapes instead of raw control bytes) so `llama.cpp` and Ollama structured-output / tool-call grammar parsers never reject tool definitions.
2. **Low-Watermark Frontier Refill (`server/engine.go`, `db/exploration.go`):**
   - In quiet mode, the planner still wakes on edge triggers whenever `open_intents == 0 && running_intents < workers`, keeping all parallel local-LLM continuous-batching slots saturated instead of idling until the heartbeat timer.
3. **Transient Unified-KV Overflow Recovery (`server/engine.go`):**
   - If concurrent slots momentarily exhaust a local server's `--kv-unified` pool (`Context size has been exceeded`), the worker returns the intent to `open` instead of permanently marking it `blocked`.
4. **Environment-Driven Context Window (`ARTEX_LLM_CONTEXT_WINDOW_K`):**
   - Configurable in `.env` and `docker-compose.yml` (default recommended: `36` for 3–4 parallel agents sharing an 80K unified KV pool) so ARTEX compacts cleanly before hitting local VRAM limits.
5. **Local-LLM-First Skills & Knowledge Pack:**
   - Added concise, file-first skills (`auth-session-testing`, `js-bundle-mining`, `cloud-staging-exposure`, `client-side-browser`) designed to keep large responses out of 36K–64K contexts.
   - Extended embedded `knowledge_lookup` with **`owasp-api`** (OWASP API Security Top 10:2023) alongside `attack`, `wstg`, and `cwe`.

---

## Verified `llama-server` Recipe (RTX 4090 24 GB + Qwen3.8-27B UD-Q4_K_XL)

Measured against Ollama 0.32.14 on the same RTX 4090 (890.5 GB/s measured copy bandwidth):

| Metric | Ollama Baseline (`-np 1`) | Tuned `llama-server` (MTP `d=2`, `-np 4`) | Uplift |
|---|---:|---:|---:|
| Single-stream decode (`c1`) | 46.6 tok/s | **79.8 tok/s** | **+71% (1.71×)** |
| 2 concurrent streams (`c2`) | 45.9 tok/s *(serial)* | **127.3 tok/s** | **+177% (2.77×)** |
| 4 concurrent streams (`c4`) | 45.9 tok/s *(serial)* | **179.3 tok/s** | **+291% (3.91×)** |
| Short-prompt TTFT | 296 ms | **157 ms** | **1.89× faster** |
| Warm 16K prefix-cache turn | — | **1.03 s** *(vs 7.02 s cold)* | **6.8× faster** |

### Recommended Launch Flags
```powershell
$env:GGML_CUDA_GRAPH_OPT = "1"
llama-server `
  -m <path-to-qwen3.8-27b-ud-q4_k_xl.gguf> `
  --alias "qwen3.8-27b-abliterated" `
  -ngl 999 -fa on `
  --ctx-size 81920 --parallel 4 --kv-unified `
  -ctk q8_0 -ctv q8_0 `
  --batch-size 2048 --ubatch-size 512 `
  --spec-type draft-mtp --spec-draft-n-max 2 `
  --ctx-checkpoints 16 `
  --reasoning-budget 1024 `
  --metrics --jinja `
  --host 0.0.0.0 --port 8080
```

### Recommended ARTEX Settings
- **`.env`:**
  ```env
  ARTEX_LLM_PROVIDER=ollama
  ARTEX_LLM_MODEL=qwen3.8-27b-abliterated
  ARTEX_LLM_BASE_URL=http://host.docker.internal:8080/v1
  ARTEX_LLM_CONTEXT_WINDOW_K=36
  ```
- **Parallel workers:** Set **`workers = 3`** in ARTEX Settings (`3 workers + 1 planner = 4 concurrent slots`, matching `--parallel 4`).
- **Role-based profiles (optional):** Bind a second profile with `thinking_type = disabled` on the same `:8080` endpoint to **Report writer** and **Goal Decomposer**, reserving full reasoning for **Planner**, **Worker**, and **Vulnerability Retest**.
