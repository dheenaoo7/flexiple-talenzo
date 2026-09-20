# Talent Search — Rubric-ARM

A recruiter-facing talent search tool over a 48-profile engineering dataset. A
recruiter describes a role in plain language; the app drafts structured
filters and a weighted scoring rubric with an LLM, runs a deterministic
filter over the dataset, and has the LLM score and explain the surviving
candidates with evidence pulled from their actual profile data. The recruiter
can then edit the filters/rubric directly, or give feedback in chat and via
per-profile fit/not-fit tags — each round produces a new, versioned state, so
nothing is silently overwritten. When satisfied, the recruiter freezes a
read-only shortlist.

## Product approach

**The core belief driving this design: recruiters won't trust a ranked list
they can't audit.** Every decision below follows from that.

- **Filters and rubric are shown and editable before the search runs, not
  buried.** The LLM's first move is to translate a fuzzy query into something
  concrete and inspectable — required/nice-to-have/excluded skills, years of
  experience, locations, company types, and a weighted list of scoring
  criteria. The recruiter sees this *before* any candidates are scored, and
  can correct it (e.g. remove a skill the LLM wrongly inferred as required)
  before spending an LLM call on scoring against a rubric they don't agree
  with.

- **Every match ships with evidence, not just a score.** Each result shows
  which rubric criteria it hit and a short quote/paraphrase tied to real
  profile fields (skills, past companies, summary). A number alone ("82/100")
  is not persuasive to a recruiter who has to defend the shortlist to a
  hiring manager; a sentence quoting the candidate's actual experience is.

- **Feedback is a conversation with visible consequences, not a black box.**
  The recruiter can react per-profile (fit / not fit) and/or type free-text
  feedback ("candidate 2 is too junior, prioritize backend depth over
  breadth"). The app responds with a plain-English `change_summary` — what it
  changed in the filters/rubric and why — before showing the re-ranked
  results. This is what makes iteration *visibly* responsive rather than a
  re-roll of the dice.

- **State is versioned, not mutated.** Every run (initial, manual filter
  edit, or chat-driven evolution) creates a new version with its own
  snapshot of filters, rubric, and results. Older versions stay browsable
  read-only via a version strip — a recruiter can always see what changed
  between round 1 and round 3, and nothing is destroyed by a bad edit.

- **Explicit states for the moments that make or break trust in an
  AI-assisted tool**: a real staged "thinking" indicator (not a spinner) so
  the recruiter knows what's happening; an empty-results state that names
  *which* filter was too strict and offers a one-click way to relax it,
  instead of a dead end; an error state (LLM failure, quota exhaustion,
  network issue) that explains what went wrong without discarding the
  recruiter's typed query or edits; and a frozen state that is unambiguously
  read-only so a finalized shortlist can't be accidentally changed later.

- **Liking a piece of evidence is a soft signal, not a hard filter.** A
  recruiter can star a specific matched-criterion bullet on a profile to
  mean "more of this, please" — it's fed back into the LLM as context on the
  next evolve round, but it never silently becomes a rigid filter the
  recruiter didn't explicitly ask for.

## Technical approach

### Architecture

```
Browser (React/Vite/TS, Tailwind)
   │  SSE for the 3 pipeline calls, plain JSON for everything else
   ▼
Go HTTP server (net/http, no framework)
   ├─ api/       handlers, SSE streaming, request/response DTOs
   ├─ session/   orchestration — ties search + llm + store together
   ├─ search/    deterministic in-memory filter over the profile dataset
   ├─ llm/       Gemini client, prompts, structured-output schemas, retries
   ├─ store/     SQLite persistence (sessions, versions, chat, liked signals)
   ├─ domain/    shared types
   └─ data/      profiles.json loaded once into memory at startup
```

### Why a deterministic filter *and* an LLM, not just an LLM

Filtering 48 profiles by hard constraints (required skills, years of
experience, location, company type) is a correctness problem, not a judgment
problem — a Go loop does it exactly and for free. The LLM is reserved for the
part that actually needs judgment: turning a fuzzy query into structured
criteria, and then scoring/explaining the *already-narrowed* candidate set
against a weighted rubric with evidence. This keeps LLM calls small (at most
20 candidate profiles per scoring call, capped and ranked by skill overlap if
more survive the filter), keeps cost/latency predictable, and means a
mis-filtered candidate can never be "rescued" by a lucky LLM guess — the
deterministic layer is the source of truth for who's even in the running.

### LLM interaction

- `google.golang.org/genai` against `gemini-2.5-flash` (model configurable via
  `GEMINI_MODEL`), every call using `ResponseMIMEType: application/json` and
  an explicit `ResponseSchema` — no ad-hoc prompt-and-hope JSON parsing.
- Three prompts, one per pipeline stage: **draft** (query → filters +
  rubric), **score** (rubric + query + up to 20 candidates + any liked
  signals → per-profile score, matched points, summary), **evolve** (current
  filters/rubric + chat text and/or fit/not-fit reactions + liked signals →
  updated filters/rubric + a `change_summary` explaining the diff in plain
  English).
- Each call has a context timeout and one automatic retry on a transient or
  malformed-JSON response. Errors are classified, not just surfaced
  verbatim: quota exhaustion (`RESOURCE_EXHAUSTED` / HTTP 429) is reported as
  non-retryable with a message that says *rate limit*, not a generic
  "connection" error — retrying a burned-out daily quota just wastes the next
  request. Unavailable/timeout errors are reported as retryable.
- Evidence is grounded, not trusted blindly: matched-point evidence text is
  expected to reference real profile fields; the app is built to drop
  unverifiable points rather than render an invented claim as fact.

### State across refinement rounds

Every pipeline call (`/run`, `/evolve`) writes a full versioned snapshot —
`filters_json`, `rubric_json`, `results_json`, and (for evolve) a
`change_summary` — rather than a diff. This trades a little storage for a lot
of simplicity and auditability: reconstructing "what did round 2 look like"
is a single row read, not a replay. `liked_signals` and `chat_messages` are
stored per-session and referenced by `version_id`, so the LLM's context for
an evolve call is always "everything the recruiter has told this session, in
order" — not just the last message.

### Streaming and failure semantics

All three pipeline endpoints (`create session`, `run`, `evolve`) stream
staged progress over Server-Sent Events, so "thinking" reflects real backend
stages, not a fake timer. On failure, the SSE stream emits a typed error
event and **no new version is written** — the session's current state is
untouched, and the frontend never discards what the recruiter had typed or
edited.

### Frontend

React + Vite + TypeScript + Tailwind, one state machine in `App.tsx`
(`landing → thinking → reviewing → results → frozen`, plus a non-destructive
error overlay). Key pieces: `FiltersRubricPanel` (editable filters + a
searchable multi-select per enum-like field, and a rubric list with an
animated weight scale), `ResultsGrid`/`ProfileCard` (evidence bullets, soft
like, fit/not-fit tags), `ChatPanel` (message history + pending reactions),
`VersionHistoryStrip` (browse prior versions read-only), `EmptyResultsState`
and `ErrorBanner` (explicit, actionable, non-destructive), `FrozenSummary`
(fully read-only final view), and `SessionHistorySidebar` (reopen past
sessions). The dev server proxies `/api` to the Go backend; the Docker setup
below does the same via nginx so the same frontend code runs unmodified in
both.

## Running it

### Option A — Docker (recommended, one command)

Requires Docker Desktop (or another Docker Engine) running.

```
cp .env.example .env
# edit .env and set GEMINI_API_KEY=<your key>
docker compose up --build
```

Then open **http://localhost:5173**. The frontend container serves the built
app via nginx and proxies `/api` to the backend container; the backend
persists its SQLite database to a named Docker volume so data survives
restarts (`docker compose down` keeps it, `docker compose down -v` wipes it).

### Option B — Run locally without Docker

```
cd backend
cp .env.example .env
# edit .env and set GEMINI_API_KEY=<your key>
go run ./cmd/server            # http://localhost:8080
```

```
cd frontend
npm install
npm run dev                    # http://localhost:5173, proxies /api to :8080
```

`.env` fields (backend):

```
GEMINI_API_KEY=       # required — get one at https://aistudio.google.com/apikey
GEMINI_MODEL=gemini-2.5-flash
PORT=8080
DB_PATH=./talent_search.db
```

The 48-profile sample dataset loads from `backend/data/profiles.json` at
startup. A SQLite file is created on first run to persist sessions, versions,
chat messages, and liked signals.

## Using it

1. Land on the search box, describe a role in plain language.
2. Watch the staged "thinking" progress while the LLM drafts filters and a
   weighted rubric.
3. Edit anything in the filters/rubric panel — including searchable
   multi-selects for skills, titles, locations, and company types — then run
   the search.
4. Review the top matches, each with evidence-grounded "why this matched"
   bullets tied to real profile fields. Star a bullet to mark it as a soft
   signal for future re-scoring; tag a profile fit/not-fit.
5. Send feedback in chat (free text and/or the pending fit/not-fit tags). The
   app explains what it changed in the filters/rubric and why, creates a new
   version, and re-runs the search. Older versions stay viewable (read-only)
   via the version strip.
6. Freeze the search for a read-only final state: frozen filters, frozen
   rubric, ranked shortlist. Start a new search any time from the sidebar.

Errors (LLM failures, quota limits, network issues) surface as a
dismissible/retryable banner and never discard in-progress work — typed
queries, edited filters, and existing versions are preserved.

## Verification

```
cd backend && go build ./... && go vet ./... && go test ./...
cd frontend && npm run build   # runs tsc -b + vite build
```

## Project layout

```
backend/
  cmd/server/main.go
  internal/{config,data,domain,store,llm,search,session,api}/
  data/profiles.json
  Dockerfile
frontend/
  src/{api,components,types.ts,App.tsx,main.tsx}
  Dockerfile
  nginx.conf
docker-compose.yml
```
