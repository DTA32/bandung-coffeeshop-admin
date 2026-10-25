# bandung-coffeeshop-admin

BDGCafe mission control (`mission-control.bdgcafe.com`): the single-admin console for editing cafe data and
reviewing what `bandung-coffeeshop-scraper` detects (field change proposals, Instagram posts, scrape sources).
It talks directly to the same PostgreSQL database as `bandung-coffeeshop-be`; **the schema and migrations are
owned by `bandung-coffeeshop-be/migrations`**, not this repo.

## Stack

GOTTH: Go 1.26 stdlib `net/http` (1.22+ pattern routing), [templ](https://templ.guide), Tailwind v4
(standalone CLI, no Node), HTMX 2. Database access is pgx/v5 with raw SQL. HTMX, Leaflet, leaflet-geoman (polygon
editor, 2.20.2, MIT) and fonts are vendored in `static/vendor/`.

## Layout

Organised by entity (`cafe`, `location`, `tag`, `rating_type`, `rating_category`, `feed`, `change`, `post`,
`source`, `user`, `review`, `dashboard`) inside each layer; handlers call controllers, controllers call the repository:

```
main.go          wiring, graceful shutdown
initializers/    env config, DB pool
routes/          <entity>.route.go: route registration per entity (+ basic.route.go for health/static, routes.go)
handlers/        shared code only: handler.go (Handler + deps), helpers.go (render/error), forms.go (parse primitives),
                 gethealth.go
handlers/<entity>/  one package per entity (cafe, location, tag, ratingtype, ratingcategory, feed, change, post,
                 source, user, review, dashboard): one struct per route, one file per route <get|post><entity><action>.go,
                 plus handler.go (shared helpers) and form.go (form parsing)
controllers/     <entity>.controller.go: business rules, validation, normalisation (no SQL)
repository/      <entity>.repository.go: all SQL (pgx) and every transaction
storage/         Cloudflare R2 client for image uploads
models/          <entity>.model.go: DTOs shared by all layers
middleware/      SameOrigin, Logging, Recover, SecurityHeaders
templates/       templ components, one file per entity (cafes.templ, rating_types.templ, ...) plus layout/components
                 (*_templ.go are generated and committed)
utils/           request/value helpers
static/          embedded CSS, JS, vendored libs
```

## Highlights

- Location form: area / district geometries are edited on a map (leaflet-geoman) bound to the GeoJSON textarea.
- Cafe form: a "Touch" button on each saved, visible tag bumps `cafe_tag.updated_at` (the public site's cafe
  "remark" is the most recently updated visible tag) without saving the form or touching `cafe.updated_at`.
- Cafe form: "About this cafe" (EN / ID) text; reviews and ratings live on `/reviews` now.
- Reviews (`/reviews`): moderation queue with pending / approved / rejected / deleted tabs, cafe and
  partnership-only filters; approve, reject, pin / unpin and fix a review's ratings. Partnership reviews (no user
  account) are created, edited and removed here and published straight away. Needs BE migration 011 (and 012
  right after deploy).
- Opening hours (BCF-19): the cafe form edits per-day hours (`cafe_hour`): an every-day default, mon–sun with
  split ranges and a "Closed" toggle, and special dates (holidays, events). Scraped hours proposals show current
  vs proposed per date and, when approved, replace the regular hours and the special dates inside the scraped
  window. Needs BE migration 013 (014 drops the old `cafe.open_hour/close_hour` after rollout).
- Images: the cafe and location forms upload JPEG / PNG / WebP files straight to Cloudflare R2 at
  `cafe/<slug>/<file>` / `location/<slug>/<file>` and fill in the `https://image.bdgcafe.com/...` URL; typing a
  URL by hand still works. Needs the `R2_*` settings (otherwise the upload buttons are hidden).
- Cafe list: sort by least / most recently updated (`?sort=updated_asc|updated_desc`).
- Lists load 50 rows at a time by infinite scroll (HTMX `revealed` sentinel; `?page=N` on the same URL, filters
  carried along; the `/changes` and `/posts` queues page by an `after=` cursor so approvals don't skip rows). Cafe and location search is server-side typeahead (`?q=`, case-insensitive match on name / id).
  Shared parsing lives in `utils/pagination.go`, SQL paging helpers in `repository/paging.go`.
- Changes: tick several proposals (or "select all loaded") and approve / reject them in one go from the sticky
  action bar; informational fields are skipped on approve, failures stay in the list with their reason.

## Local setup

```bash
cp .env.example .env     # set DB_USER, DB_PASSWORD, DB_HOST, ...
make run                 # templ generate + tailwind build + go run .
```

Config: `APP_PORT`, `DB_HOST/PORT/USER/PASSWORD/NAME`, `DB_SSLMODE`, `IMAGE_BASE_URL`, `DISPLAY_TZ`
(default `Asia/Jakarta`), `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET` (image uploads;
leave blank to disable). The database must already be migrated by `bandung-coffeeshop-be`.

Other targets: `make generate` (templ), `make css` (Tailwind; downloads the CLI to `bin/` on first use),
`make build`, `make test`, `make vet`, `make fmt`. Run `make generate` after editing `.templ` files and
`make css` after adding Tailwind classes.

## Tests

```bash
make test        # unit tests, -race
# DB integration tests (repository) run only against a migrated database:
ADMIN_TEST_DSN='postgres://postgres:postgres@localhost:5432/bandung_coffeeshop?sslmode=disable' \
  go test ./repository -race
```

## Docker

```bash
docker build -t bandung-coffeeshop-admin .
```

The image runs `templ generate` and the Tailwind build, then compiles the root `main` package.

## Deployment

CI is Woodpecker (`.woodpecker/prod.yml`) and publishes to `ghcr.io/dta32/bandung-coffeeshop-admin/prod`;
the cluster is managed in flux-gitops. The app has **no auth of its own**: Envoy basic auth
(`SecurityPolicy bdgcafe-mission-control-basicauth`) guards `mission-control.bdgcafe.com`, so never expose the
Service any other way. `middleware.SameOrigin` blocks cross-site POSTs.
