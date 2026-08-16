# Kylistran (web)

Angular frontend for Kylistran — a private novel-reading and writing site. Reading and editing both talk to the API in `../api`; see the repo root [README](../README.md) for the overall architecture and deployment.

## Reading

`/library` lists books and chapters and renders them, fetched live from the API. It's protected by a lightweight shared-password gate meant for friends/family, not real access control — see "Security model" below.

## Writing

`/editor` is the author-only chapter editor, gated separately by a real login (JWT from the API, not the shared reader password). It has three views:

- **Books** (`/editor`) — create, rename, reorder, and delete books.
- **Chapters** (`/editor/:bookId`) — create, reorder, and delete chapters within a book.
- **Chapter editor** (`/editor/:bookId/:chapterId`) — a single textarea where each non-blank line becomes one paragraph. Saving is immediately live for readers — there's no separate publish step or rebuild.

To insert an image, type a marker line anywhere in the text (or use the "Insert image at cursor" form, which does this for you):

```
{{image src="/images/example.png" alt="Description for accessibility" caption="Optional caption"}}
```

Image *files* aren't uploaded through the editor — `src` must point at something already deployed under `public/images/`. Adding a genuinely new image file still means copying it into `public/images/<path>` and running a normal `npm run deploy` from this directory; only chapter text/structure is fully dynamic.

Music links (shown as a small section below the chapter text) are managed as their own list in the editor, not part of the textarea.

## Changing the reader password

The shared password is checked client-side as a SHA-256 hash, stored in `src/app/core/auth/access-config.ts`. It ships in the public JS bundle as a hash, not plaintext, so pick something longer/less common than a dictionary word — the hash is unsalted and could be brute-forced offline by someone motivated enough.

To generate a new hash, run this in any browser's devtools console:

```js
crypto.subtle.digest('SHA-256', new TextEncoder().encode('your-new-password'))
  .then(b => console.log(Array.from(new Uint8Array(b)).map(x => x.toString(16).padStart(2, '0')).join('')))
```

Paste the resulting hex string into `ACCESS_PASSWORD_HASH_HEX` in `access-config.ts`.

## Security model — what this does and doesn't protect against

The **reader** gate is intentionally not a real access-control system:

- It's client-side only. It stops casual visitors and most bots, but a determined person could read the compiled JavaScript and find ways around it.
- `robots.txt`, a `noindex` meta tag, and an `X-Robots-Tag` HTTP header ask search engines and AI crawlers not to index or archive the site. Well-behaved crawlers respect this; not all of them do.
- Images are served as regular static files under `/images/...` and are reachable by anyone with the direct URL, gate or no gate.

The **editor** login is real auth (bcrypt-hashed password, JWT, server-side enforcement on every write) — it's a different, stronger mechanism from the reader gate above, and the two are independent.

## Local development

```sh
npm install
npm start
```

Requires the API running locally (see the repo root README) — `src/environments/environment.ts` points at `http://localhost:8081` by default. Open `http://localhost:4200/`.

## Deploying

```sh
npm run deploy
```

Builds the app and runs `firebase deploy --only hosting`. One-time setup: `npx firebase login`, then a Firebase project ID in `.firebaserc`.
