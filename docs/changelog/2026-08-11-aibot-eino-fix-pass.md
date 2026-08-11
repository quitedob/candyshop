# Changelog / Devlog — 2026-08-11 aibot (Eino) Fix Pass

**Date:** 2026-08-11
**Scope:** `backend/internal/pkg/eino/` + AI handlers (Go 1.24 / Gin / Eino v0.7.36), `frontend/pages/admin/ai/` (Nuxt 3 / Vue 3 / TS)
**Trigger:** Deep review of the "aibot" (Eino agent) workflow — verified it uses Eino correctly (ChatModelAgent / DeepAgent / planexecute / compose.Graph / adk streaming), then fixed every confirmed defect found.
**Result:** 23 files changed, **+145 / −816**. Backend `go build` + `go vet` + `go test ./...` green; frontend `nuxi typecheck` green for the changed page; `go mod tidy` consistent. All findings fixed; 3 non-blocking caveats documented (§5).

---

## 1. Process

1. **Learn from devlog first** — three parallel read-only Explore agents read `docs/changelog/2026-08-11-code-review-fixes.md` (fix-pass conventions + known AI/Eino debt), the Eino reference docs (`docs/research/eino/`), and mapped the entire aibot flow (frontend → route → handler → `pkg/eino` → tools → graph → SSE).
2. **Verify before reporting** — every finding from the scan was hand-checked against source (`sse_handler.go`, `pipeline.go`, `plan_order.go`, `ai.go`, `index.vue`) before dispatch.
3. **Fix with parallel subagents** — 6 general-purpose agents, each owning a *disjoint file set* so concurrent edits could not collide (table §2).
4. **Lead runs final merged verification** — `go build` / `go vet` / `go test` / `go mod tidy -diff` since concurrent agents can't see each other's edits.
5. **Review subagent** — an independent general-purpose agent re-verified every fix against source + the Eino v0.7.36 API and reported a per-fix PASS/FAIL table + regressions.

---

## 2. How subagents were used (parallel, non-conflicting)

| Agent | Ownership (files only it touched) | Fixes |
|-------|-----------------------------------|-------|
| A | `pkg/eino/graph/pipeline.go`, `pkg/eino/plan_order.go`, `handlers/system/handler.go` | #3 CI ref underflow, #5 dead persister param |
| B | `handlers/system/sse_handler.go` | #2/#4 doc-type mapping + nested if, #6 temperature wiring |
| C | `handlers/admin/admin_trade_ai.go`, `cmd/api/main.go` | #6 admin temperature, #7 loud fallback |
| D | `frontend/pages/admin/ai/index.vue` | #1 generate-quotation payload, #6 temperature/model params |
| E | `pkg/eino/tool/` (+ `go.mod`/`go.sum` via tidy) | #8 vendor ReviewEditTool, #9 delete 9 dead tools |
| F | `config/ai_config.go`, `pkg/eino/client.go`, `pkg/eino/agent.go` | #10 RAG config flag, legacy-runner cleanup |

Each agent was given the exact finding with `file:line` references, told to **read before edit**, **match surrounding style**, and **run `go build ./...`** before reporting. After all six finished, the lead ran the merged verification, then dispatched an independent **review subagent** to check final status (it re-ran build/vet/test/typecheck and cross-checked each fix against the Eino API).

---

## 3. Fixes in detail

### Broken in production

#### 1. Admin `generate-quotation` mode always returned 400
- **Problem:** the AI console POSTed `{ requirements: text }` but the backend `GenerateQuotation` binds `prompt` / `customerRequirements` (`handlers/system/ai.go:145`) → the key was ignored and every call returned `prompt_or_requirements_required`.
- **Fix:** `frontend/pages/admin/ai/index.vue` now sends `{ customerRequirements: text }`. Mode works end-to-end.

#### 2. Customer portal never refreshed docs after AI generation
- **Problem:** the registered trade-document tool is the graph tool `generate_trade_documents` (`graph/tool.go:33`), but the system SSE switch mapped only legacy names (`generate_proforma_invoice`, …). The admin processor handled `generate_trade_documents` (`admin_trade_ai.go:257`), the system one did not — so `processAgentEvent` never emitted `document_type` and the customer trade page (`[id].vue:326`) never called `fetchAllDocs()`.
- **Fix:** `sse_handler.go:302-307` adds a `generate_trade_documents` case that parses the `doc_types` array from `msg.ToolCalls[0].Function.Arguments` via a new `firstDocTypeFromArgs` helper (`:357`) and emits the first recognized type, falling back to `"TRADE_DOCUMENT"` so the refresh trigger always fires.

#### 3. Commercial Invoice reference underflow
- **Problem:** `pipeline.go:202` `piRef := fmt.Sprintf("PI-%s-%05d", ts, seq-1)` — when `seq := now.UnixMilli() % 100000` is `0`, `%05d` renders `-0001` (negative doc reference). It also referenced a PI number one below the actual one.
- **Fix:** clamp `seq` to `≥1` (`pipeline.go:185-189`) and use the same `seq` for `piRef` (`:207`).

### Correctness / honesty

#### 4. Redundant nested `if`
- Removed the duplicated `if len(msg.ToolCalls) > 0` in `processAgentEvent` (`sse_handler.go:275`).

#### 5. Dead `DocumentPersister` param + stale docs in Plan-Execute-Replan
- `NewOrderProcessingAgent` accepted `_ einotool.DocumentPersister` it never used; the doc comment claimed the executor "generate documents" though it binds only `check_compliance` / `validate_lc_documents` / `track_shipment`.
- **Fix:** dropped the param (`plan_order.go:37`, caller `handler.go:162`) and corrected the comments (`plan_order.go:10,33`).

#### 6. Model/temperature UI was cosmetic
- The console showed `model: 'MiniMax-M2.7', temperature: 0.7` and sent neither. Backend agents ran at OpenAI defaults.
- **Fix:** temperature is now real — the 3 system SSE handlers + the admin handler parse an optional `temperature` (0–2) query param and pass `adk.WithChatModelOptions([]model.Option{model.WithTemperature(t)})` to `runner.Query` (Eino v0.7.36 API). The frontend appends `&temperature=` and `&model=`. A `model` mismatch logs a warning (per-request model override is not supported; the server keeps `OPENAI_MODEL`).

#### 7. Admin silently dropped streaming
- `AdminAITradeChat` fell back to non-streaming `Generate` when `tradeAgent == nil`, but `InitAgent` failure was only a bare warning.
- **Fix:** the fallback now logs `Error:` loudly (`admin_trade_ai.go:136,143`) and the `main.go:133` warning states the consequence ("admin trade chat will run in non-streaming mode"). Still non-fatal (AI is an optional dependency).

### Dependency / dead code

#### 8. Eino examples repo as a production dependency
- `track_shipment.go`, `translate_content.go`, `validate_lc_documents.go` imported `github.com/cloudwego/eino-examples/adk/common/tool` (a pseudo-versioned example repo) for `InvokableReviewEditTool`.
- **Fix:** vendored the wrapper into `pkg/eino/tool/review_edit.go` (Apache-2.0, byte-identical to upstream; `schema.Register[*ReviewEditInfo]()` + stateful-interrupt/resume logic intact) and switched the 3 tools to the local type. `go mod tidy` removed `eino-examples` from `go.mod`/`go.sum`; zero references remain.

#### 9. Nine dead `generate_*.go` tools
- `generate_pi`, `generate_ci`, `generate_sales_contract`, `generate_packing_list`, `generate_certificate_of_origin`, `generate_health_certificate_request`, `generate_ingredients_declaration`, `generate_insurance_certificate_request`, `generate_shipper_letter_of_instruction` were never registered anywhere (superseded by the `generate_trade_documents` graph tool) and carried a `json.Marshal`-of-string content bug.
- **Fix:** verified zero callers (grep on every exported constructor), then deleted all 9. `persister.go` (`DocumentPersister` interface) retained — still used by `agent.go` / `deep_b2b.go` / `graph/pipeline.go`.

#### 10. Hardcoded RAG toggle + dead legacy-runner conditional
- `client.go:30` `const RAGComplianceEnabled = false` was hardcoded; the `buildClientRunner` compliance-tool conditional was dead (always `nil`).
- **Fix:** `AIConfig.RAGComplianceEnabled` now loads from `AI_RAG_COMPLIANCE_ENABLED` (default false, `config/ai_config.go:22,42`); `client.go` uses `var RAGComplianceEnabled` set in `NewClient`, building `complianceRetriever` when enabled; `agent.go` still registers `compliance_lookup` when enabled. `buildClientRunner` simplified to `(ctx, chatModel)` — dead conditional and always-nil checkpoint-store param removed.

---

## 4. Verification

| Check | Result |
|-------|--------|
| `go build ./...` (backend) | exit 0 |
| `go vet ./...` (backend) | exit 0 |
| `go test ./...` (backend) | pass, 0 FAIL |
| `go mod tidy -diff` | no changes proposed (eino-examples removed, deps consistent) |
| `npx nuxi typecheck` (frontend) | no errors in `pages/admin/ai/index.vue` (pre-existing errors elsewhere untouched) |
| Review subagent | 10/10 fixes PASS, no regressions |

---

## 5. Deliberately NOT changed (documented caveats)

- **Draft-doc PI/CI cross-reference semantics** — `generateDocument` recomputes `time.Now()`/`seq` per doc type, so a CI generated in the same batch still may not reference the exact PI number. The numbering scheme needs a trade-doc redesign, not a bug fix.
- **Model-mismatch warning log spam** — the admin console always sends `&model=MiniMax-M2.7`; if the server `OPENAI_MODEL` differs, every SSE request logs a Warning. Harmless; sync the frontend default to the server env to silence.
- **Admin `generate_trade_documents` mapping** — `admin_trade_ai.go` still maps it unconditionally to `PROFORMA_INVOICE` while the system handler parses `doc_types`. Pre-existing inconsistency; the admin UI doesn't consume `document_type`.
