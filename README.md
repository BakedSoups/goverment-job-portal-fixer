# Bay Area Gov Jobs

Bay Area Gov Jobs is a free, server-rendered search tool for public-sector jobs around the Bay Area.

I tried applying to the City and County of San Francisco to see what software engineers can do in government, but found the existing job portal difficult to navigate. I built this tool to make the listings easier to search and compare. It became useful enough that I made the source public and the website free to use.

## How it works

At startup, the Go application requests JSON from these public endpoints:

```text
GET https://api.smartrecruiters.com/v1/companies/CityAndCountyOfSanFrancisco1/postings?limit=100&offset={offset}
GET {ref URL returned by each SmartRecruiters posting summary}
GET https://www.governmentjobs.com/careers/{tenant}/home/loadJobsOnMaps
```

The GovernmentJobs tenant values cover the configured Bay Area cities and counties, including `contracosta`, `sanmateo`, `santaclara`, `marin`, `sonoma`, `napacounty`, `solanocounty`, `alamedaca`, `berkeley`, `fremontca`, `haywardca`, `oaklandca`, `richmondca`, `mountainview`, `paloaltoca`, `sanjoseca`, `cityofsantaclaraca`, `sunnyvale`, and `vallejo`.

Each JSON response is decoded into provider-specific Go structs and then normalized into a common job model. The normalized jobs are indexed in memory; searches and live filters use that index, so changing a filter does not request the endpoints again. The current application does not write a local JSON cache or database.

The explainable search engine uses familiar techniques instead of a black-box model:

- tokenization and word-frequency counts;
- regular expressions for structured signals such as required experience;
- a taxonomy of canonical skills and aliases;
- title, department, skill, and description matches for ranking.

The interface uses Go templates, small progressive-enhancement JavaScript, Leaflet, OpenStreetMap tiles, and U.S. Census Bureau county boundaries.

## Run locally

Requires Go 1.22 or newer and network access to the public job sources.

```sh
go run .
```

Open <http://localhost:8080>.

## Zero deployment

Zero runs Cloudflare Workers rather than Go processes. The deployment adapter under `deploy/zero/` serves a normalized snapshot produced by the Go scraper and search index:

```sh
go run ./cmd/export-snapshot -output /tmp/jobs.json
curl http://localhost:8080/ -o /tmp/index.html
node deploy/zero/build-payload.mjs --index /tmp/index.html --snapshot /tmp/jobs.json --output /tmp/zero-deploy.json
```

The generated payload contains public job data and static assets only. It does not contain credentials.

Zero custom domains currently work only for domains purchased through Zero's domain service. Domains bought from another registrar need an HTTPS reverse proxy in front of the generated `*.app.withzero.ai` address; a DNS-only CNAME does not establish Zero's required hostname and TLS mapping.

## Checks

```sh
sh scripts/pr-check.sh
```

The checker runs formatting, tests, template validation, and repository checks.
