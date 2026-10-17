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
`source`, `dashboard`) inside each layer; handlers call controllers, controllers call the repository:

```
main.go          wiring, graceful shutdown
initializers/    env config, DB pool
routes/          <entity>.route.go: route registration per entity (+ basic.route.go for health/static, routes.go)
handlers/        shared code only: handler.go (Handler + deps), helpers.go (render/error), forms.go (parse primitives),
                 gethealth.go
handlers/<entity>/  one package per entity (cafe, location, tag, ratingtype, ratingcategory, feed, change, post,
                 source, dashboard): one struct per route, one file per route <get|post><entity><action>.go,
                 plus handler.go (shared helpers) and form.go (form parsing)
controllers/     <entity>.controller.go: business rules, validation, normalisation (no SQL)
repository/      <entity>.repository.go: all SQL (pgx) and every transaction
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
(default `Asia/Jakarta`). The database must already be migrated by `bandung-coffeeshop-be`.

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
