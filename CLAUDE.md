# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## About

`bandung-coffeeshop-admin` is **BDGCafe mission control** (`mission-control.bdgcafe.com`, BCF-11): the
single-admin console for editing cafe data and reviewing what `bandung-coffeeshop-scraper` (BCF-12)
detects. It connects straight to the same PostgreSQL database as `bandung-coffeeshop-be`; the schema
(and every migration) is owned by `bandung-coffeeshop-be/migrations`. Plan:
`bandung-coffeeshop-be/docs/freshness-pipeline-plan.md`.

Stack: **GOTTH** — Go 1.26 stdlib `net/http` (1.22+ pattern routing), [templ](https://templ.guide),
Tailwind v4 (standalone CLI, no Node), HTMX 2. pgx/v5 for the database.

**Auth is not in this app.** Envoy basic auth (flux-gitops `SecurityPolicy`
`bdgcafe-mission-control-basicauth`, secret `bdgcafe-mission-control-basic-auth`) guards it. Never expose the Service any other way.
`middleware.SameOrigin` rejects cross-site POSTs because browsers re-send basic-auth credentials.

## Commands

```bash
make run          # templ generate + tailwind build + go run . (reads .env)
make generate     # templ -> templates/*_templ.go (generated files are committed)
make css          # static/css/input.css -> static/css/app.css (gitignored, built in Docker)
make test         # go test ./... -race -cover
go vet ./...
make fmt          # templ fmt + go fmt

# Repository integration tests run only against a migrated database (001–013):
ADMIN_TEST_DSN='postgres://postgres:postgres@localhost:5432/bandung_coffeeshop?sslmode=disable' go test ./repository

docker build -t bandung-coffeeshop-admin .
```

Copy `.env.example` → `.env`. Config: `APP_PORT`, `DB_HOST/PORT/USER/PASSWORD/NAME`, `DB_SSLMODE`,
`IMAGE_BASE_URL` (renders R2 keys), `DISPLAY_TZ` (default `Asia/Jakarta`).

After editing a `.templ` file run `make generate`; after adding Tailwind classes run `make css`.
Pin versions: templ `v0.3.1020` (also in `go.mod`, Dockerfile, Woodpecker), Tailwind `v4.3.3`
(Makefile + checksum in Dockerfile), HTMX 2.0.11, Leaflet 1.9.4 and
`@geoman-io/leaflet-geoman-free` 2.20.2 (MIT, `static/vendor/leaflet-geoman/`, no CDN at runtime) vendored in `static/vendor/`.

## Architecture

Layers: `handlers` -> `controllers` -> `repository` -> PostgreSQL. SQL lives only in `repository/`;
business rules (validation, normalisation, proposal/feed rules) live only in `controllers/`; handlers parse
requests and render templates.

```
main.go          wiring: repository.New(pool) -> controllers.New(repos) -> handlers.New(ctl, cfg) -> routes.New
                 (layout follows Sinanaas/gotth-auction)
initializers/    LoadEnv/LoadConfig (env config; DSN pins the session timezone to UTC), ConnectDB (pgx pool)
routes/          <entity>.route.go (cafe, location, tag, rating_type, rating_category, feed, change, post, source,
                 user, review, dashboard) + basic.route.go (health, static) + routes.go (middleware chain)
handlers/        shared code only, package `handlers`: handler.go (Handler{Ctl, Cfg}), helpers.go (Render/ServerError/
                 NotFound/UserError/SavedOr/FriendlyUnique), forms.go (Opt*/Checked/ParseImages/FormErrorf), gethealth.go
handlers/<entity>/  one package per entity (cafe, location, tag, ratingtype, ratingcategory, feed, change, post, source,
                 user, review, dashboard); each route is one small struct embedding *handlers.Handler, one file per route named
                 <get|post><entity><action>.go (cafe/getcafes.go, cafe/postcafeedit.go, change/postchangeapproveall.go,
                 source/getsourcesnapshot.go), plus handler.go (helpers used by several routes of the entity) and
                 form.go (+ form_test.go) for form parsing. Cross-entity helpers are exported (cafe.FromPath,
                 feed.ParseFeedForm); routes/ imports the packages
controllers/     <entity>.controller.go: XxxController with validation + business rules, calls the repository;
                 controllers.go (Controllers bundle), errors.go (ValidationError, IsValidation, IsNotFound, shared
                 validators)
repository/      <entity>.repository.go: XxxRepository, all raw pgx SQL; repository.go (Repositories bundle,
                 sentinel errors). Owns every transaction. Integration tests (gated on ADMIN_TEST_DSN) sit here
models/          <entity>.model.go DTOs shared by all layers
middleware/      SameOrigin (CSRF guard), Logging, Recover, SecurityHeaders, CacheStatic
templates/       templ components, one file per entity (cafes, locations, tags, rating_types, rating_categories,
                 feeds, changes, posts, sources, users, reviews, dashboard) + layout/components/icons .templ, helpers.go, styles.go
utils/           request helpers (PathInt, QueryInt, IsHTMX, RedirectOK) and small value helpers
static/          embedded (embed.go): css/, js/forms.js (Leaflet picker + geoman polygon editor, image list, toggles), vendor/ (htmx, leaflet, fonts)
```

Entities: `cafe`, `location` (non-cafe poi/area/district + `location_image`), `tag`, `rating_type`
(`rating_type_label`), `rating_category`, `feed` (`cafe_feed`), `change` (field `change_proposal`s), `post`
(`ig_post` proposals), `source` (`source_target` + `scrape_run` + snapshots), `user` (`users`, BE migration 009,
BCF-25), `review` (`cafe_review` + its `cafe_rating` rows, BE migrations 011 + 012, BCF-31), `dashboard` (home counts incl. active / deleted users and pending reviews; its controller also reads `SourceRepository.LastRuns`). Naming convention everywhere: `<entity>.<layer>.go`
(`cafe.model.go`, `cafe.repository.go`, `cafe.controller.go`, `cafe.route.go`); templates use the plural
(`cafes.templ`). `change` and `post` share the `change_proposal` table, so the lock/resolve/get/reject helpers
are package functions in `repository/change.repository.go`.

Transactions (all in `repository/`): `CafeRepository.Create/Update` (location + cafe + price/hours/tags/images),
`ReviewRepository.SetStatus/SetPinned` (row locked `FOR UPDATE`), `ReviewRepository.ReplaceRatings`,
`ReviewRepository.CreatePartnership/UpdatePartnership` (review + ratings), `LocationRepository.Create/Update` (location + images), `TagRepository.Delete` (cafe_tag + tag),
`ChangeRepository.Apply` (live column or windowed `cafe_hour` replace + proposal resolve + `cafe.updated_at`/`location.updated_at` bump),
`ChangeRepository.Reject` / `PostRepository.Ignore`, `PostRepository.CreateFeed` (cafe_feed insert + proposal
approve). Controllers never open a transaction.

There is no `toast/`: feedback is the existing `?ok=` redirect message and HTMX fragments.

Concrete types, no interfaces between layers (matching the backend). Pure logic (form parsing,
`controllers.Normalize*`, `FieldUpdateFor`, `BuildFeed`) is unit tested next to its controller/handler; SQL is
covered by the gated `repository` integration tests.

### Pages

- `/` dashboard (pending changes / posts / reviews, failing sources, active / deleted users, last run per source)
- `/cafes`, `/cafes/new`, `/cafes/{location_id}` — full cafe form (basics, details incl. weather / phone /
  menu / status, per-day opening hours + a 7-day preview, Leaflet coordinates, "About this cafe" EN / ID = `cafe.description(_indo)` (empty → NULL; the
  address stays in `location.description(_indo)`), pricing, tags, ordered images); soft delete. Reviews and
  ratings are not on the cafe form any more (see `/reviews`). The list
  sorts by last update (`?sort=updated_asc` default / `updated_desc`, whitelisted in `CafeRepository.List`) and
  searches server-side (`?q=`, typeahead).
  Each saved + visible tag row has a "Touch" button (`POST /cafes/{id}/tags/{tagId}/touch`, HTMX, swaps the row's
  touch fragment) that sets only `cafe_tag.updated_at = now()` (not `cafe.updated_at`) so that tag becomes the
  public "remark"; the newest visible tag is marked "current remark"
- `/cafes/{id}/feed` — `cafe_feed` list, create, edit, hide/unhide, expiry
- `/locations`, `/locations/new`, `/locations/{id}` — poi / area / district with GeoJSON; a map is bound two-way to the
  GeoJSON textarea: poi = point picker, area / district = leaflet-geoman polygon editor (draw / edit / drag / cut /
  remove; Polygon or MultiPolygon; blur the textarea to redraw, invalid JSON shows an inline error)
- `/tags`, `/rating-types` (`rating_type_label` dimensions), `/rating-categories` — inline HTMX CRUD
- `/changes` — pending field proposals grouped by cafe: approve / edit & approve / reject, or select several
  rows and batch approve / reject;
  informational fields (`temp_closed`, `menu_content`, `ig_bio`) are dismiss-only; "approve all"
- `/posts` — pending `ig_post` proposals: "Create feed for this cafe" (prefilled form) or "Ignore"
- `/sources` — `source_target` per cafe × source (enable, external id override/clear, reset failures),
  recent `scrape_run`s, `/snapshots/{id}` raw/parsed view
- `/users` — public-site accounts (BCF-27): typeahead search on name / email (ILIKE via `EscapeLike`), status
  tabs all / active / deleted, newest first, infinite scroll like every list (see below). The typeahead and tabs
  target `#user-list` and get the `UserResults` fragment (tabs + table); the scroll sentinel gets `UserRows`. `POST /users/{id}/delete` / `revive` only set `users.status` (no-op if unchanged) and
  return the row. The BE checks `users.status` on every request (deleted = logged out at once) and hides a deleted
  user's reviews at read time, so `/users` never touches Redis sessions or reviews
- `/reviews` — review moderation (BCF-31): status tabs pending (default) / approved / rejected / deleted, cafe
  filter (`?cafe=`), partnership-only toggle (`?partnership=1`); reviews of deleted users are left out. Pending
  is FIFO (`updated_at ASC, id ASC`), the other tabs newest first; cursor paging `after=<last review id>`.
  Card actions (HTMX, swap the card; `tab` is sent so a card that leaves the tab renders compact):
  `POST /reviews/{id}/approve` (pending / rejected → approved), `/reject` (pending / approved → rejected),
  `/pin`, `/unpin` (any non-deleted review). `GET/POST /reviews/{id}/ratings` edits a review's ratings.
  `GET/POST /reviews/new`, `GET/POST /reviews/{id}/edit` (partnership reviews only, else 404) and
  `POST /reviews/{id}/remove` (partnership only → `deleted`; HTMX from the queue, form post from the edit page)

### List pages: pagination and typeahead

Every list page loads 50 rows at a time (`repository.PageSize`) by infinite scroll. Repositories fetch
`LIMIT PageSize+1 OFFSET (page-1)*PageSize` through `pageWindow` / `toPage` (`repository/paging.go`) and return
`models.Page[T]{Items, HasMore}`; every ORDER BY ends in a unique tiebreaker (id). Handlers call
`utils.ParseListParams(r, listID)` (`page`, `q`, `Fragment`) and `lp.Next(path, params, hasMore)` to build the
next-page URL, which must carry every active filter (q, sort, type, cafe, source, failing...). A request is a
*fragment* request (GET + `HX-Request`, not an htmx history restore, and `page >= 2` or `HX-Target` == the list
container id) and then renders only the rows component; otherwise the full page. Each list has a templ rows
component (`CafeListRows`, `TargetRows`, ...) used by both the full page and the fragment, ending in
`@LoadMore(next)` (or `LoadMoreRow` in tables): a sentinel with `hx-trigger="revealed" hx-swap="outerHTML"`,
absent on the last page. Cafes and locations also search server-side: `?q=` is matched with `ILIKE` on
name / id, `%`, `_` and `\` escaped by `repository.EscapeLike`; the input is a typeahead (`hx-get`,
`input changed delay:300ms, search`, `hx-push-url`, `hx-include` for sort/type). The approval queues
page by cursor instead of offset, because approving rows already on screen shrinks the queue and an offset
would then skip unseen rows: `/changes` takes `after=<last cafe id>` (50 cafe groups per page, ordered by
location name, id), `/posts` takes `after=<last proposal id>` (ordered by taken_at desc, id desc) and
`/reviews` takes `after=<last review id>` (keyset on updated_at, id; ascending on the pending tab); `page`
still increments so fragment detection works. Lists needing everything keep an unpaged method (`CafeRepository.Options`,
`TagRepository.List`, `RatingTypeRepository.List`, `RatingCategoryRepository.List`).

`/changes` also has multi-select batch actions (`POST /changes/batch`, `action=approve|reject`, repeated `id=`,
max `controllers.MaxBatch` = 200): `ChangeController.ApplyBatch` runs each id through the single-row
Get/Approve/Reject path, skips informational fields on approve, never aborts on one failure, and the HTMX
response removes processed rows and fills `#batch-flash` via out-of-band swaps. Checkbox / action-bar logic is
delegated JS in `static/js/forms.js`.

### UI

The catalog pages (home, cafes, locations, tags, rating types/categories) are a 1:1 port of the old
React console (`bandung-coffeeshop-fe-console`, `src/routes/admin.*` + `src/components/admin/*`):
same markup, Tailwind classes, labels and section order. The shared class strings live in
`templates/styles.go` (`cardCls`, `labelCls`, `inputCls`, button classes) — reuse them instead of adding
component CSS. Icons are inline lucide SVGs (`templates/icons.templ`). Theme tokens are copied verbatim
from `bandung-coffeeshop-fe/src/styles.css`; the font is the vendored Plus Jakarta Sans Variable.
`forms.js` is loaded once by `Layout` and replaces the console's React state (rating /
tag toggles, `data-confirm`, "Saving..." labels). Saving a cafe/location returns to its list.

### Proposal rules (keep in sync with the scraper)

Approving a field writes the live column, marks the proposal `approved` with `resolved_value`, and
sets `cafe.updated_at = now()` (plus `location.updated_at` for `address` / `status`) — one transaction
(`ChangeRepository.Apply` in `repository/change.repository.go`; the field -> column validation is `controllers.FieldUpdateFor`). Field → column: `hours` → `cafe_hour` rows (see "Opening hours" below), `phone` →
`cafe.phone`, `menu` → `cafe.menu` (approval requires an http(s) URL; only the cafe form also accepts `#image-N` gallery hashes), `address` → `location.description` (the address lives there;
`description_indo` is never touched), `status` → `location.status`. Reject/dismiss/ignore =
`rejected`; the scraper never re-proposes a rejected value. A post approval inserts `cafe_feed` and
stores `{"cafe_feed_id": n}`.

### Opening hours (BCF-19; schema = BE migration 013)

A `cafe_hour` row has `day` (`day_enum` mon..sun, null = every day), `open_hour` / `close_hour`,
`valid_from` / `valid_until` (null on both = regular hours, otherwise a dated "special" row, `valid_until`
inclusive; a row with one bound is still dated) and `close`. Rules (shared with be and scraper; the BE
CLAUDE.md "Opening hours" is the source): for date D use the first tier that has rows — (a) dated rows
covering D, a matching day beating a null day; (b) undated rows for D's weekday; (c) undated null-day rows.
Several rows in a tier = split hours; a closed row wins in its tier; no row / null hours = unknown;
close < open = overnight; 24h = `00:00–23:59`; dates are Jakarta (`models.Jakarta`, fixed UTC+7).
`models.ResolveDay` / `ResolveDays` (`models/cafe_hour.model.go`) port `bandung-coffeeshop-be`
`service/hours.go` — keep them in sync.

- Cafe form: `templates.HoursEditor` (every-day row, mon..sun, each with ranges + "Closed", "Add another time
  range", and a list of special dates from / until (blank = same day) / optional weekday). It posts the
  parallel lists `hours_group` / `hours_open` / `hours_close` (one per range row), `hours_closed` (closed group
  keys) and `special_key` / `special_from` / `special_until` / `special_day`; group keys are `all`, mon..sun or
  a special key. `handlers.ParseHours` turns them into rows, `controllers.NormalizeHours` validates (HH:MM, open
  + close both set, no hours on a closed row, not closed and open for the same day / dates, YYYY-MM-DD,
  from <= until) and sorts. Saving deletes and re-inserts all the cafe's rows in the cafe transaction
  (`replaceAllHours`). JS (add / remove, "Closed" disables the ranges) is delegated in `forms.js`.
- `hours` proposal `new_value` (scraper): `{"window_from", "window_until", "rows": [{day, open, close,
  valid_from, valid_until, closed}], "days": [raw scraped days]}` (`models.HoursProposal`); the queue's
  current value is `{"rows": [...]}`, the live rows in that shape. The queue shows a current-vs-proposed table
  per date of the window (changed days highlighted, special days marked). Approve (and "Edit & approve", the
  same editor prefilled with the proposed rows; the window is fixed) replaces the undated rows plus the dated
  rows overlapping `[window_from, window_until]` and keeps dated rows outside it (`models.ReplacedByWindow`,
  `replaceWindowHours`), bumping `cafe.updated_at` in the same transaction; proposed dated rows must overlap
  the window. A proposal that does not parse in this shape (the old `{open, close, days}`) shows its raw JSON
  and can only be rejected.
- Deploy order: BE migration 013 → this admin (+ be and scraper on `cafe_hour`) → 014 (drops
  `cafe.open_hour/close_hour`). This admin fails on a database without 013.

### Review rules (BCF-31; schema = BE migrations 011 + 012)

- "Visible" on the public site = `status = 'approved' AND (user_id IS NULL OR users.status = 'active')`.
  `user_id IS NULL` = partnership review (admin-entered, `reviewer_name` only).
- User reviews: the admin may only approve / reject, pin / unpin and edit the ratings. Text, scores, visit month,
  link and `updated_at` stay as the author wrote them. The public BE puts an edited review back to `pending`
  itself; this app never needs to. User reviews are never deleted here (the author does that; deleting the user
  on `/users` hides them).
- Moderation and rating edits never bump `cafe_review.updated_at` (= the author's last edit, which orders the
  public list). Status / pin changes lock the row with `SELECT ... FOR UPDATE` and check the allowed source
  statuses (`controllers.reviewTransitions`); a lost race returns `ErrReviewStateChanged`, shown on the card.
- Partnership reviews: cafe, reviewer name and content required; overall / WFC score optional, 1–5 in 0.1 steps;
  visit month optional (`<input type="month">` → first of the month); external link optional, http(s) only, a
  title needs a link; pin checkbox. Created `approved`; editing sets `updated_at = now()` and keeps the status;
  "Remove" sets `deleted`.
- Ratings belong to a review (`cafe_rating.review_id`) and are validated against the type's bucket bounds
  (`RatingTypeRepository.Bounds`, `controllers.ValidateRatings`). Saving replaces them: `DELETE ... WHERE
  review_id = $1` then `INSERT` (cafe_id, category_type, score, overrides, review_id) in one transaction. Never
  use `ON CONFLICT`: until 012 the table is `UNIQUE (cafe_id, category_type)`, after 012 it is
  `UNIQUE (review_id, category_type)`, so no conflict target works on both. Before 012 a second review rating
  the same type of the same cafe fails with `ErrRatingTaken` (shown to the admin).
- Deploy order: BE migration 011 → this admin (+ the BE) → 012. Never reference `cafe_review.is_subjective`,
  `personal_score` or `content_indo` (dropped by 012).

### Data conventions

- Timestamps are `timestamp without time zone` in UTC; the UI shows and accepts `DISPLAY_TZ`.
- Opening hours: `cafe_hour` rows (BE migration 013; `cafe.open_hour/close_hour` are never read or written
  here and are dropped by 014). See "Opening hours" below.
- `cafe_feed.images` and `location_image.url` hold full `https://image.bdgcafe.com/...` URLs;
  scraped Instagram media lives under R2 `ig/` (watermark bypassed).
- `cafe.weather` is NOT NULL — always pass a non-nil slice.
- Saving a cafe only bumps `cafe_tag.updated_at` when visibility changes (the public API picks the
  most recently updated tag as the remark).
