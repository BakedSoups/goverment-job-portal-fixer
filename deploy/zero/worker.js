let snapshotPromise;

export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    if (url.pathname === "/health") {
      const snapshot = await loadSnapshot(env, request.url);
      return Response.json({ ok: true, jobs: snapshot.documents.length, generatedAt: snapshot.generatedAt });
    }

    if (url.pathname.startsWith("/jobs/")) {
      const snapshot = await loadSnapshot(env, request.url);
      const id = decodeURIComponent(url.pathname.slice(6));
      const document = snapshot.documents.find((candidate) => candidate.Job.ID === id);
      if (!document) return new Response("Not found", { status: 404 });
      return htmlResponse(renderJobPage(document, snapshot.tags));
    }

    if (url.pathname === "/" && url.search) {
      const snapshot = await loadSnapshot(env, request.url);
      return htmlResponse(renderFilteredPage(snapshot, url.searchParams));
    }

    if (url.pathname === "/") {
      return env.ASSETS.fetch(new URL("/shell.html", request.url));
    }

    return env.ASSETS.fetch(request);
  },
};

async function loadSnapshot(env, requestURL) {
  if (!snapshotPromise) {
    const assetURL = new URL("/data/jobs.json", requestURL);
    snapshotPromise = env.ASSETS.fetch(assetURL).then(async (response) => {
      if (!response.ok) throw new Error(`snapshot returned ${response.status}`);
      return response.json();
    });
  }
  return snapshotPromise;
}

function renderFilteredPage(snapshot, params) {
  const query = (params.get("q") || "").trim();
  const selectedTags = splitValues(params.getAll("tags"));
  const selectedGovernments = splitValues(params.getAll("gov"));
  let minimum = clampNumber(params.get("yoe_min"), 0);
  let maximum = clampNumber(params.get("yoe_max"), 10);
  if (minimum > maximum) [minimum, maximum] = [maximum, minimum];

  const tagMap = new Map(snapshot.tags.map((tag) => [tag.id, tag]));
  const terms = selectedTags.length > 0 ? selectedTags : canonicalTerms(query, snapshot.tags);
  const active = new Set(terms);
  const allowedGovernments = new Set(selectedGovernments);

  const results = [];
  for (const document of snapshot.documents) {
    const job = document.Job;
    if (allowedGovernments.size > 0 && !allowedGovernments.has(job.SourceID)) continue;
    if (!experienceMatches(job, minimum, maximum)) continue;
    const ranked = scoreDocument(document, terms);
    if (terms.length > 0 && ranked.score === 0) continue;
    results.push({ document, score: ranked.score || 1, reasons: ranked.reasons });
  }

  results.sort((a, b) => b.score - a.score || Date.parse(b.document.Job.ReleasedDate) - Date.parse(a.document.Job.ReleasedDate));
  const districts = mapDistricts(results, selectedGovernments, snapshot.documents);
  const cards = results.map((result) => renderJobCard(result, active, tagMap, params)).join("");
  const body = cards || '<p class="empty">No jobs matched this query.</p>';

  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Bay Area Gov Jobs</title></head><body>
    <div class="results-column"><section class="results-header"><h2>${results.length} matching jobs</h2></section><section class="results-list">${body}</section></div>
    <div data-leaflet-map data-districts="${escapeAttribute(JSON.stringify(districts))}"></div>
  </body></html>`;
}

function scoreDocument(document, terms) {
  if (terms.length === 0) return { score: 1, reasons: [] };
  let score = 0;
  const reasons = [];
  const title = document.Job.Title.toLowerCase();
  const department = document.Job.Department.toLowerCase();

  for (const term of terms) {
    let termScore = 0;
    const conceptCount = document.ConceptHits?.[term] || 0;
    const frequencyCount = document.Frequencies?.[term] || 0;
    if (conceptCount > 0) {
      termScore += conceptCount * 12;
      reasons.push(`${term} concept matched`);
    }
    if (frequencyCount > 0) {
      termScore += frequencyCount * 3;
      reasons.push(`${term} frequency matched`);
    }
    const plainTerm = term.replaceAll("_", " ");
    if (title.includes(plainTerm) || title.includes(term)) {
      termScore += 30;
      reasons.push(`${term} matched title`);
    }
    if (department.includes(term)) {
      termScore += 8;
      reasons.push(`${term} matched department`);
    }
    if (termScore === 0) return { score: 0, reasons: [] };
    score += termScore;
  }
  return { score, reasons: [...new Set(reasons)] };
}

function canonicalTerms(query, tags) {
  if (!query) return [];
  const lower = query.toLowerCase();
  const aliases = new Map();
  for (const tag of tags) {
    aliases.set(tag.id.toLowerCase(), tag.id);
    aliases.set(tag.label.toLowerCase(), tag.id);
    for (const alias of tag.aliases || []) aliases.set(alias.toLowerCase(), tag.id);
  }

  const terms = [];
  const seen = new Set();
  for (const [alias, id] of aliases) {
    if (/[^a-z0-9_]/.test(alias) && lower.includes(alias) && !seen.has(id)) {
      terms.push(id);
      seen.add(id);
    }
  }
  for (const token of lower.match(/[a-z0-9+#.-]+/g) || []) {
    const id = aliases.get(token) || token;
    if (!seen.has(id)) {
      terms.push(id);
      seen.add(id);
    }
  }
  return terms.sort();
}

function experienceMatches(job, minimum, maximum) {
  if (!job.RequiredYOEFound) return minimum === 0;
  const requiredMaximum = Math.max(job.RequiredYOEMax, job.RequiredYOEMin);
  return job.RequiredYOEMin <= maximum && requiredMaximum >= minimum;
}

function mapDistricts(results, selectedGovernments, documents) {
  const counts = new Map();
  for (const result of results) counts.set(result.document.Job.SourceRegion, (counts.get(result.document.Job.SourceRegion) || 0) + 1);
  const max = Math.max(0, ...counts.values());
  const sourceRegions = new Map(documents.map((document) => [document.Job.SourceID, document.Job.SourceRegion]));
  const selectedRegions = new Set(selectedGovernments.map((id) => sourceRegions.get(id)).filter(Boolean));
  const regions = [["north-bay", "North Bay"], ["sf", "SF"], ["east-bay", "East Bay"], ["peninsula", "Peninsula"], ["south-bay", "South Bay"]];
  return regions.map(([id, name]) => {
    const count = counts.get(name) || 0;
    let level = 0;
    if (count > 0 && max > 0) {
      if (count * 4 >= max * 3) level = 4;
      else if (count * 2 >= max) level = 3;
      else if (count * 4 >= max) level = 2;
      else level = 1;
    }
    return { id, name, count, level, selected: selectedRegions.has(name) };
  });
}

function renderJobCard(result, active, tagMap, params) {
  const job = result.document.Job;
  const tags = (result.document.ConceptNames || []).map((id) => tagMap.get(id)).filter(Boolean).sort((a, b) => {
    const activeDifference = Number(active.has(b.id)) - Number(active.has(a.id));
    return activeDifference || a.label.localeCompare(b.label);
  });
  const tagHTML = tags.map((tag) => `<span class="${active.has(tag.id) ? "is-matched" : ""}">${escapeHTML(tag.label)}</span>`).join("");
  const titleMatch = result.reasons.some((reason) => reason.includes("matched title"));
  const hasCriteria = active.size > 0;
  const label = !hasCriteria ? "" : titleMatch ? "Title match" : result.score >= 30 ? "Strong match" : "Related match";
  const detail = label === "Title match" ? "Your search terms appear in the job title." : label === "Strong match" ? "Several relevant skills or terms appear in the listing." : label ? "Relevant skills or terms appear in the listing." : "";
  const search = new URLSearchParams(params).toString();
  const experience = job.RequiredYOEFound ? `Required: ${job.RequiredYOEMin} ${plural(job.RequiredYOEMin, "year", "years")}` : "Required experience: not specified";
  const salary = job.SalaryMin ? `<small>${money(job.SalaryMin)} - ${money(job.SalaryMax)}</small>` : "";

  return `<article class="job-card" data-region="${escapeAttribute(job.SourceRegion)}"><div>
    <p class="eyebrow">${escapeHTML(job.Department)}</p>
    <div class="source-badges"><span class="source-badge source-${escapeAttribute(job.SourceID)}">${escapeHTML(job.SourceName)}</span><span class="source-region">${escapeHTML(job.SourceRegion)}</span></div>
    <h2><a href="/jobs/${encodeURIComponent(job.ID)}${search ? `?${escapeAttribute(search)}` : ""}">${escapeHTML(job.Title)}</a></h2>
    <p class="summary"><span>${escapeHTML(job.Location)}</span><span>${escapeHTML(job.Employment)}</span><span>${formatDate(job.ReleasedDate)}</span><span>${experience}</span></p>
    <div class="tags">${tagHTML}</div></div><aside>${label ? `<strong class="match-label" title="${escapeAttribute(detail)}">${label}</strong><span class="match-detail">${detail}</span>` : ""}${salary}</aside></article>`;
}

function renderJobPage(document, tags) {
  const job = document.Job;
  const tagMap = new Map(tags.map((tag) => [tag.id, tag]));
  const signals = (document.ConceptNames || []).map((id) => tagMap.get(id)).filter(Boolean);
  const signalButtons = signals.map((tag) => `<button type="button" data-signal-tag="${escapeAttribute(tag.id)}" data-aliases="${escapeAttribute(JSON.stringify(tag.aliases || []))}">${escapeHTML(tag.label)}</button>`).join("");
  const sections = (job.Sections || []).filter((section) => section.Text).map((section) => `<section class="job-section"><h2>${escapeHTML(section.Title)}</h2><div class="listing-lines">${section.Text.split("\n").map((line) => `<span data-listing-line>${escapeHTML(line)}</span>`).join("")}</div></section>`).join("");
  const required = job.RequiredYOEFound ? `${job.RequiredYOEMin} ${plural(job.RequiredYOEMin, "year", "years")} · ${escapeHTML(job.RequiredYOEConfidence)} confidence` : "not specified";
  const requiredButton = job.RequiredYOEFound ? `<button type="button" data-signal-tag="required_yoe" data-highlight-text="${escapeAttribute(job.RequiredYOESource)}">Required ${job.RequiredYOEMin} ${plural(job.RequiredYOEMin, "year", "years")}</button>` : "";
  const preferredButton = job.PreferredYOEFound ? `<button type="button" data-signal-tag="preferred_yoe" data-highlight-text="${escapeAttribute(job.PreferredYOESource)}">Preferred ${job.PreferredYOEMin}-${job.PreferredYOEMax} years</button>` : "";

  return pageShell(job.Title, `<article class="job-detail" data-region="${escapeAttribute(job.SourceRegion)}">
    <nav class="back"><a href="/">Back to search</a></nav><header class="job-heading"><p class="eyebrow">${escapeHTML(job.Department)}</p>
    <div class="source-badges"><span class="source-badge source-${escapeAttribute(job.SourceID)}">${escapeHTML(job.SourceName)}</span><span class="source-region">${escapeHTML(job.SourceRegion)}</span></div>
    <h1>${escapeHTML(job.Title)}</h1><div class="meta-grid"><span>${escapeHTML(job.Location)}</span><span>${escapeHTML(job.Employment)}</span><span>${escapeHTML(job.Experience)}</span><span>${formatDate(job.ReleasedDate)}</span><span>${escapeHTML(job.RefNumber)}</span></div>
    <p class="actions"><a class="button" href="${escapeAttribute(job.ApplyURL)}">Apply</a><a class="button secondary" href="${escapeAttribute(job.PostingURL)}">Original listing</a></p></header>
    <section class="matched-tags"><h2>Skills &amp; requirements found</h2><p class="signal-help">Select a skill or requirement to highlight the evidence in the original listing.</p><p class="parsed-fact"><strong>Required experience:</strong> ${required}</p><div class="tags signal-tags">${requiredButton}${preferredButton}${signalButtons || "<span>No taxonomy matches yet</span>"}</div></section>${sections}</article>`);
}

function pageShell(title, content) {
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${escapeHTML(title)}</title><link rel="stylesheet" href="/static/app.css"><script src="/static/app.js" defer></script></head><body><header class="topbar"><a class="brand" href="/"><span>Bay Area Gov Jobs</span></a><div class="topbar-actions"><span class="topbar-meta">Job information provided through public APIs</span><button class="theme-toggle" type="button" data-theme-toggle aria-pressed="false">Dark mode</button><a class="github-link" href="https://github.com/BakedSoups/goverment-job-portal-fixer" target="_blank" rel="noopener noreferrer">GitHub</a></div></header><main>${content}</main></body></html>`;
}

function splitValues(values) {
  return [...new Set(values.flatMap((value) => value.split(",")).map((value) => value.trim()).filter(Boolean))];
}

function clampNumber(value, fallback) {
  const number = Number.parseInt(value ?? "", 10);
  return Number.isFinite(number) ? Math.max(0, Math.min(10, number)) : fallback;
}

function formatDate(value) {
  if (!value || value.startsWith("0001-")) return "";
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric", year: "numeric", timeZone: "UTC" }).format(new Date(value));
}

function money(value) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 }).format(value);
}

function plural(number, singular, pluralForm) {
  return number === 1 ? singular : pluralForm;
}

function escapeHTML(value) {
  return String(value ?? "").replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&#39;");
}

function escapeAttribute(value) {
  return escapeHTML(value).replaceAll("`", "&#96;");
}

function htmlResponse(html) {
  return new Response(html, { headers: { "Content-Type": "text/html; charset=utf-8", "Cache-Control": "no-store" } });
}
