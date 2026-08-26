# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Kylistran: a private writing app (one author, occasionally shared with friends/family) — Angular reader/editor frontend + Go API for storing chapters. Monorepo: `web/` (Angular 19) and `api/` (Go stdlib `net/http`), no shared code between them beyond the wire format.

## Commands

**API** (from `api/`):
- `go run ./cmd/server` — run locally (port `8081` by default, loads `api/.env` automatically, avoids clashing with an unrelated `workout-api` on `8080`)
- `go build ./...` / `go vet ./...`
- No test files exist yet.

**Web** (from `web/`):
- `npm start` — dev server at `http://localhost:4200`, pointed at `http://localhost:8081` (`environment.ts`)
- `npm run build`
- `npm run deploy` — `ng build --configuration production && firebase deploy --only hosting`
- `npm test` — Karma/Jasmine configured but no spec files currently exist
- No lint config (no ESLint/Prettier) in either project.

**Deploy** (see README.md for full details): API via `docker compose up -d --build` (must pass `--build` — compose won't rebuild the Go binary on its own); web via `npm run deploy` from `web/`. `docker-compose.yml` only defines the `kylistran-api` service — the frontend isn't part of it.

## Architecture

### Two independent auth systems — don't confuse them
- `web/src/app/core/auth/` — a shared-password **reader gate** in front of `/library` (`GateComponent` at `/`, `authGuard`). Not real auth; just keeps the site from being fully public.
- `web/src/app/core/editor-auth/` — real JWT auth for `/editor` (`editor-auth.guard.ts`, `auth.interceptor.ts` attaches the bearer token). Backed by `api/internal/auth/` (`middleware.go`: `RequireAuth` validates the JWT, `RequireAdmin` checks the `admin` claim). Single seeded author account only — no signup; seeded from `ADMIN_USERNAME`/`ADMIN_PASSWORD` on first API startup only when `users` table is empty (`api/internal/db/db.go` `SeedAdmin`).

### API route split (`api/internal/handlers/handlers.go` `Register()`)
- **Public, unauthenticated**: `GET /books`, `GET /books/{slug}`, `GET /books/{slug}/chapters/{slug}` — these must stay filterable-but-directly-accessible (e.g. a book's list endpoint filters hidden/unpublished items, but its direct-by-slug endpoint intentionally does not).
- **Admin, JWT + admin claim required**: everything under `/admin/*` (books and chapters CRUD/reorder), wrapped via the local `admin(fn)` helper (`RequireAuth(RequireAdmin(fn))`), not a route-group middleware.
- Handlers for both public and admin variants of a resource live in the same file (e.g. `book_handlers.go` has `listBooks`/`getBookDetail` alongside `listAdminBooks`/`createBook`/`updateBook`/`deleteBook`) — there's no separate admin package.
- Admin-only DTOs (`models.AdminBookSummary`, `AdminChapterSummary`, `AdminChapter`) deliberately expose the numeric `id` that public DTOs never do.

### Data model & schema
`books` → `chapters` → `chapter_blocks` / `chapter_music_links`, each with an explicit `position` column (book order, chapter order, block order). Schema: `api/internal/db/schema.sql`.

**No migration framework** — schema is applied via `CREATE TABLE IF NOT EXISTS` on every startup (`db.go` `Open()`), which does nothing for columns added to an already-existing table. Any new column needs a matching `ALTER TABLE ... ADD COLUMN` in `db.go`'s `migrate()` function (guarded by ignoring the sqlite "duplicate column name" error, since `migrate()` also runs against fresh DBs that already have the column from `schema.sql`).

`chapter_blocks.block_type` is `'paragraph'` or `'image'`. `models.Block` (`api/internal/models/models.go`) has hand-written `MarshalJSON`/`UnmarshalJSON` so the wire format matches the frontend's `ChapterBlock = string | ChapterImage` union exactly (a plain JSON string vs. an object) — no separate DTO translation layer. A paragraph string starting with `#` is an author-only comment: persisted and editable normally, but skipped by the public reader's render loop (`chapter-reader.component.ts` `isAuthorComment()`) — it is not filtered anywhere in the API or the editor.

### Frontend structure
Standalone components (no NgModules), Angular signals, new `@if`/`@for` control-flow syntax. Feature folders under `web/src/app/features/`:
- `gate/` — reader password gate
- `library/` — public reading UI (`book-list`, `chapter-list`, `chapter-reader`), talks to `content/content.service.ts` (public API endpoints, `content.models.ts`)
- `editor/` — author CMS (`book-manager`, `chapter-manager`, `chapter-editor`, `login`), talks to `editor-content.service.ts` (`/admin` endpoints, separate `Admin*`/`*Input` models mirroring the Go admin DTOs)
- `maintenance/` — full-site maintenance mode, toggled by `core/maintenance.ts`'s `MAINTENANCE_MODE` constant, which swaps the entire route table in `app.routes.ts`

`chapter-editor` stores chapter body as one big textarea, not per-block inputs — round-tripped through `serializeBlocks`/`parseBlocks` in `chapter-editor/block-parser.ts`. Image blocks are inline markers in that text, not separate form fields.

No shared editor layout/nav component exists — `book-manager`/`chapter-manager`/`chapter-editor` are three independent routed pages (`editor.routes.ts`) that each link to the next via plain `routerLink`s in their own templates. Editor list rows (`book-manager`, `chapter-manager`) follow a shared CSS pattern: `.editor__row` as a CSS grid with BEM-ish per-field classes (`.editor__title`, `.editor__slug`, etc.) so a `@media (max-width: 640px)` block can retarget them via `grid-template-areas` without touching markup.

### Deployment topology (see README.md for the full runbook)
API runs as a Docker container (`api/Dockerfile`, CGO-disabled Go build, `modernc.org/sqlite` is pure-Go) behind a NUC's existing Caddy reverse proxy, reachable at `/chapters/*` with the prefix stripped (the app's own routes are mounted at root — `/books`, not `/chapters/books`). It joins an external Docker network `edge` rather than merging compose projects. The frontend deploys separately to Firebase Hosting; nothing from the API is baked into the web build.
