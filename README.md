# Bay Area Gov Jobs

Bay Area Gov Jobs is a free, server-rendered search tool for public-sector jobs around the Bay Area.

I tried applying to the City and County of San Francisco to see what software engineers can do in government, but found the existing job portal difficult to navigate. I built this tool to make the listings easier to search and compare. It became useful enough that I made the source public and the website free to use.

## How it works

The Go application collects public job information from:

- the SmartRecruiters public API for San Francisco;
- public GovernmentJobs JSON endpoints for other Bay Area governments.

At startup, raw postings are normalized into a common job model and indexed in memory. Searches and live filters use that index, so changing a filter does not scrape the APIs again.

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

## Checks

```sh
sh scripts/pr-check.sh
```

The checker runs formatting, tests, template validation, and repository checks.
