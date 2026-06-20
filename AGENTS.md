# AGENTS.md

Rules for agents working in this repository.

## Project Direction

This project recreates the SF jobs portal as a Go server-rendered application.

Core goals:

- Scrape jobs from the public SmartRecruiters API.
- Normalize raw postings into stable internal Go models.
- Parse descriptions into clean text, sections, tags, and term frequencies.
- Build an explainable old-school search engine using tokenization, regex, taxonomy aliases, and frequency scoring.
- Render HTML through Go templates.
- Keep the code modular, boring, and easy to inspect.

## Expected Layout

Prefer this structure as the project grows:

```text
.
├── main.go
├── scraper/
├── jobs/
├── parser/
├── taxonomy/
├── search_engine/
├── storage/
├── web/
├── templates/
├── static/
└── scripts/
```

Package responsibilities:

- `scraper/`: remote API clients and external API response structs.
- `jobs/`: internal job model and normalization from scraper structs.
- `parser/`: HTML cleanup, text normalization, salary/deadline extraction.
- `taxonomy/`: canonical concepts and aliases such as Python, SQL, NoSQL, GIS, Power BI.
- `search_engine/`: tokenization, indexing, regex matching, frequency counts, query parsing, ranking.
- `storage/`: memory, JSON, and later SQLite persistence.
- `web/`: routes, handlers, template loading, HTTP server setup.
- `templates/`: server-rendered HTML templates only.
- `static/`: CSS and small browser-side JavaScript only.
- `scripts/`: local automation such as PR checks.

## Go Rules

- Keep `main.go` thin. It should wire dependencies, start the scrape/index flow, and launch the server.
- Do not put scraper, parser, ranking, or template rendering internals in `main.go`.
- Use typed structs at package boundaries. Avoid passing loose `map[string]any` values through the app.
- Keep external API structs separate from internal app models.
- Prefer small packages with clear ownership over a large utility package.
- Do not duplicate business logic across handlers, templates, or packages.
- If two handlers need the same behavior, move it into a package-level function or method.
- If two templates need the same markup, move it into a partial template.
- Run `gofmt` on all Go code.
- Keep errors wrapped with context.
- Use standard library first unless a dependency clearly reduces complexity.

## Template Rules

- Templates live under `templates/`.
- Shared markup goes under `templates/partials/`.
- Use `define` blocks for reusable partials.
- Do not duplicate job cards, search boxes, filters, headers, footers, or tag lists across pages.
- Keep templates focused on presentation. Do not implement parsing, ranking, or data cleanup in templates.
- Template data should be prepared by Go handlers or view-model builders.
- Prefer clear view models over exposing raw scraper structs to templates.

## Search Engine Rules

- Search must be explainable.
- Keep concept aliases in `taxonomy/`, not scattered through handlers.
- Keep regex matchers in `search_engine/regex.go` or parser-specific regex files.
- Searches like `python software engineer`, `data sql`, and `nosql analyst` should resolve through canonical concepts where possible.
- Store both raw frequency counts and matched canonical concepts.
- Ranking should be simple enough to explain in the UI.

## Data Rules

- Preserve raw job text where useful, but render cleaned and normalized text.
- Keep raw scraped data separate from normalized app data.
- Cache API responses or normalized jobs when doing repeated local development.
- Do not hard-code a single job listing into app logic.

## Web Rules

- Server-render HTML from Go.
- Avoid frontend frameworks unless explicitly requested.
- Use CSS in `static/app.css`.
- Use browser JavaScript only for progressive enhancement.
- The first screen should be the actual searchable jobs interface, not a marketing landing page.

## PR Readiness

Before considering work done, run:

```sh
sh scripts/pr-check.sh
```

The checker should remain fast and local. If a new package, dependency, or build step is added, update the checker in the same change.

## Change Discipline

- Keep changes scoped to the requested task.
- Do not refactor unrelated code while adding a feature.
- Do not remove user work.
- Add tests when changing parser, taxonomy, ranking, storage, or normalization behavior.
- For UI work, verify templates render through the Go server.

