# Talent Search — Rubric-ARM

Recruiter-facing talent search over a 48-profile dataset: free-text query →
LLM-drafted filters + weighted rubric → deterministic filter → LLM-scored,
evidence-backed matches → chat/tag feedback evolves the search (versioned) →
freeze a final shortlist.

## Product approach

- Filters + rubric are shown and editable **before** search runs — no black box.
- Every match shows evidence quoted from real profile data, not just a score.
- Feedback (chat + fit/not-fit tags) gets a plain-English "what changed and why".
- Every round is a new version, never a silent overwrite — full history browsable.
- Explicit states for thinking, empty results, errors, and frozen — none of
  them destroy in-progress work.
- Liking a piece of evidence is a soft signal (context for re-scoring), never
  a hard filter.

## Technical approach

**Stack**: Go backend (net/http, SQLite via `modernc.org/sqlite`, no cgo) +
React/Vite/TS/Tailwind frontend + Gemini (`google.golang.org/genai`).

- **Deterministic filter + LLM, not LLM-only.** Hard constraints (skills,
  years, location, company type) are a plain Go filter — exact, free, no
  hallucination risk. The LLM only scores/explains the already-narrowed set
  (capped at 20 candidates).
- **Schema-constrained JSON, not prompt-and-hope.** Every Gemini call uses
  `ResponseMIMEType: application/json` + an explicit `ResponseSchema`.
- **Three single-shot prompts**: draft (query → filters + rubric), score
  (rubric + candidates → scores + evidence), evolve (feedback → updated
  filters/rubric + change summary). No chat/agent loop — the model is
  stateless.
- **Grounded evidence, enforced in code.** Each matched-point's evidence text
  is token-matched against the candidate's real profile fields after
  generation; anything that doesn't match real data is dropped, not shown.
- **State is rebuilt from SQLite per call, not remembered by the model.**
  Every evolve round reconstructs: original query, latest filters/rubric,
  currently-shown ranked results, this round's feedback, and all liked
  signals across the session. The response becomes a new version row —
  calibration is append-only, never a mutation.
- **Failure handling**: 45s timeout + 1 retry per call, errors classified
  (quota/429 vs. unavailable vs. generic) so retry guidance is accurate. On
  any failure, no new version is written and nothing typed is lost.
- **Streaming**: all three pipeline endpoints stream real staged progress
  over SSE — not a fake spinner.

## Decisions

- **Deterministic filter vs. RAG, for candidate retrieval.** Went with a
  plain deterministic filter, not RAG. The dataset is small (48 profiles, fits
  fully in memory/context), and the LLM is already used up front to parse the
  free-text query into structured filters — so precise, exact-match filtering
  is cheap and correct with no retrieval layer needed. RAG earns its cost at
  a larger dataset size, where an in-memory filter can't hold/scan everything
  and semantic retrieval is needed to narrow the pool before scoring.
- **Versioning vs. a single chat session with state.** Went with explicit
  versions (filters + rubric + results snapshotted per round), not a chat
  thread carrying state implicitly. A chat-only approach loses structured
  filter/profile data inside free-text turns — recovering "what were the
  filters after round 2" means re-parsing chat history. Versioning stores
  that state directly and keyed for lookup, so past filters/rubric/results
  are a single row read, not a replay — far easier and more reliable to
  access than reconstructing state from a chat log.

## Running it

### Docker (one command)

Requires Docker running.

```
cp .env.example .env
# set GEMINI_API_KEY in .env
docker compose up --build
```

Open **http://localhost:5173**. SQLite persists to a named volume
(`docker compose down -v` to wipe it).

### Without Docker

```
cd backend
cp .env.example .env   # set GEMINI_API_KEY
go run ./cmd/server     # http://localhost:8080
```

```
cd frontend
npm install
npm run dev              # http://localhost:5173, proxies /api to :8080
```

`.env` fields: `GEMINI_API_KEY` (required, [get one here](https://aistudio.google.com/apikey)),
`GEMINI_MODEL` (default `gemini-2.5-flash`), `PORT`, `DB_PATH`.

## Using it

1. Describe a role in plain language.
2. Review/edit the LLM-drafted filters + weighted rubric, then run.
3. Review ranked matches with evidence bullets; star evidence or tag fit/not-fit.
4. Send feedback in chat — filters/rubric evolve with a stated reason, new version, re-ranked.
5. Freeze for a read-only final shortlist. Start a new search any time.

## Verification

```
cd backend && go build ./... && go vet ./... && go test ./...
cd frontend && npm run build
```

## Project layout

```
backend/   cmd/server, internal/{config,data,domain,store,llm,search,session,api}, data/profiles.json, Dockerfile
frontend/  src/{api,components,types.ts,App.tsx,main.tsx}, Dockerfile, nginx.conf
docker-compose.yml
```
