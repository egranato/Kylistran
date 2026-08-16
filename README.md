# Kylistran

A small, private writing app: an Angular reader/editor frontend plus a self-hosted Go API for storing chapters. Built for one author, occasionally shared with friends and family.

## Architecture

- **`web/`** — Angular 19 SPA, deployed to Firebase Hosting. Reading and editing both go through the API below; nothing is baked into the build.
- **`api/`** — Go (stdlib `net/http`), SQLite (`modernc.org/sqlite`, pure Go, no CGO), JWT auth, deployed as a Docker container behind DuskBird's existing Caddy reverse proxy.
- **Users**: single author account only — no public sign-up. The account is seeded from `ADMIN_USERNAME`/`ADMIN_PASSWORD` on first API startup (only when the `users` table is empty).
- **Reads are public**: `GET /books`, `GET /books/{slug}`, `GET /books/{slug}/chapters/{slug}` require no auth — the Angular app's own shared-password reader gate (`/library`) is what's in front of them for friends/family. **Writes require the author's JWT**: everything under `/admin/*`, gating the `/editor` route.
- **Routing**: the API is reachable at `https://kylistran.tplinkdns.com/chapters/*`, proxied by DuskBird's Caddy (path prefix stripped before reaching the Go app — the app's own routes are mounted at root, e.g. `/books`, not `/chapters/books`).
- **Networking**: the API container joins an external Docker network `edge`, which DuskBird's `caddy` service also joins (already exists from the `workout` deployment), so Caddy can reach it by container name without merging compose projects.

## One-time NUC setup

The `edge` network already exists from the `workout` deployment — nothing to create. DuskBird's `Caddyfile` needs one addition, then redeploy DuskBird to pick it up.

Add to `Caddyfile` in the DuskBird repo, alongside the existing `/workouts/*` block, **before** the catch-all `reverse_proxy api:3000`:

```caddyfile
handle_path /chapters/* {
    reverse_proxy kylistran-api:8080
}
```

Then from the DuskBird repo directory: `docker compose up -d` (only the `caddy` container restarts).

## Deploying the API

```powershell
cp .env.example .env   # then fill in JWT_SECRET, ALLOWED_ORIGINS, ADMIN_USERNAME, ADMIN_PASSWORD
docker compose up -d --build
```

`ADMIN_USERNAME`/`ADMIN_PASSWORD` only take effect on first startup, when no users exist yet. There's no way to change them later short of editing the database directly — pick a real password.

No host port is published — the API is only reachable through Caddy at `/chapters/*`, or directly container-to-container on `edge`.

## Deploying the frontend

```powershell
cd web
npm install
npm run deploy
```

(`ng build --configuration production && firebase deploy --only hosting`.) Make sure `ALLOWED_ORIGINS` on the API includes the resulting Hosting URL(s), and that `web/src/environments/environment.production.ts` points at the right API URL.

## Local development (no Docker, no Caddy)

**API:**
```powershell
cd api
cp .env.example .env   # then fill in JWT_SECRET, ADMIN_USERNAME, ADMIN_PASSWORD
go run ./cmd/server
```

`api/.env` (gitignored) is loaded automatically on startup — vars already set in the real environment take precedence, so this doesn't interfere with the Docker deployment above. Defaults to port `8081` to avoid clashing with `workout-api`'s `8080` if both run locally at once.

**Web:**
```powershell
cd web
npm install
npm start
```

`web/src/environments/environment.ts` already points at `http://localhost:8081`. Open `http://localhost:4200/`.

## Data model

`books` → `chapters` → `chapter_blocks` / `chapter_music_links`, each with an explicit `position` column since order matters (book order, chapter order within a book, block order within a chapter). Schema lives at `api/internal/db/schema.sql`, applied idempotently on API startup.

`chapter_blocks.block_type` is `'paragraph'` or `'image'` — the Go `models.Block` type has custom JSON (de)serialization so the wire format matches the frontend's `ChapterBlock = string | ChapterImage` union exactly, with no translation layer.

## Verification checklist

1. `docker network inspect edge` — both DuskBird's Caddy container and `kylistran-api` attached.
2. `curl -i https://kylistran.tplinkdns.com/` — DuskBird's root route unaffected.
3. `curl -i https://kylistran.tplinkdns.com/chapters/healthz` — `200 ok` from the Go app.
4. `curl -i https://kylistran.tplinkdns.com/chapters/books` — returns the seeded book list.
5. Log in to the deployed editor (`https://<hosting-url>/editor/login`) as the seeded author account, confirm the book/chapter list loads, edit and save a chapter, and confirm it appears immediately on the public reader page — no rebuild needed.
