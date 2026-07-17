(function () {
  onReady(() => {
    safeInit(initLeafletMap, "Map");
    safeInit(initTagSearch, "Search");
    safeInit(initGovernmentPicker, "Government filter");
    safeInit(initLiveFilters, "Live filters");
    safeInit(initSignalHighlights, "Highlights");
    safeInit(initGitHubStars, "GitHub stars");
  });
})();

function onReady(callback) {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", callback, { once: true });
    return;
  }
  callback();
}

async function initGitHubStars() {
  const output = document.querySelector("[data-github-star-count]");
  if (!output) return;

  const response = await fetch("https://api.github.com/repos/BakedSoups/goverment-job-portal-fixer", {
    headers: { Accept: "application/vnd.github+json" },
  });
  if (!response.ok) return;

  const repository = await response.json();
  if (!Number.isInteger(repository.stargazers_count)) return;
  output.textContent = repository.stargazers_count.toLocaleString();
  output.setAttribute("aria-label", `${repository.stargazers_count.toLocaleString()} GitHub stars`);
}

function safeInit(init, label) {
  function report(error) {
    console.error(`${label} failed to initialize`, error);
    if (label === "Map") {
      const status = document.querySelector("[data-map-status]");
      if (status) status.textContent = `${label} failed: ${error.message}`;
    }
  }

  try {
    const result = init();
    if (result && typeof result.catch === "function") result.catch(report);
  } catch (error) {
    report(error);
  }
}

async function initLeafletMap() {
  const holder = document.querySelector("[data-leaflet-map]");
  if (!holder) return;

  const canvas = holder.querySelector("[data-leaflet-canvas]");
  const status = holder.querySelector("[data-map-status]");
  function setStatus(message) {
    if (status) status.textContent = message;
  }

  function showMapError(message) {
    canvas.innerHTML = `<div class="map-error">${message}</div>`;
    setStatus(message);
  }

  if (!window.L) {
    showMapError("Leaflet did not load.");
    return;
  }

  let districts = [];
  let points = [];
  try {
    districts = JSON.parse(holder.dataset.districts || "[]");
    points = JSON.parse(holder.dataset.points || "[]");
  } catch {
    showMapError("Map data could not be parsed.");
    return;
  }
  setStatus("Loading official county boundaries...");

  const regionsResponse = await fetch(holder.dataset.regionsUrl);
  if (!regionsResponse.ok) {
    showMapError("Bay Area boundary data could not be loaded.");
    return;
  }
  const regions = await regionsResponse.json();
  if (!regions || regions.type !== "FeatureCollection") {
    showMapError("Bay Area boundary data is invalid.");
    return;
  }

  const map = L.map(canvas, {
    attributionControl: false,
    scrollWheelZoom: true,
    zoomControl: false,
  }).setView([37.82, -122.18], 8);

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    errorTileUrl: "data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs=",
  }).addTo(map);
  L.control.zoom({ position: "topright" }).addTo(map);
  L.control.attribution({ position: "bottomright", prefix: false })
    .addAttribution('&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>')
    .addAttribution('Boundaries: <a href="https://www.census.gov/geographies/mapping-files/time-series/geo/cartographic-boundary.html">U.S. Census Bureau</a>')
    .addTo(map);

  const rootStyle = getComputedStyle(document.documentElement);
  const regionColor = (id) => {
    return rootStyle.getPropertyValue(`--region-${id}`).trim() || "#667067";
  };
  let districtByID = new Map(districts.map((district) => [district.id, district]));
  const districtStyle = (feature) => {
    const district = districtByID.get(feature.properties.id);
    const color = regionColor(feature.properties.id);
    return {
      color: district && district.selected ? color : "#fffdf8",
      fillColor: color,
      fillOpacity: 0.38,
      opacity: 0.95,
      weight: district && district.selected ? 4 : 1.5,
    };
  };
  const districtLayer = L.geoJSON(regions, {
    style: districtStyle,
    interactive: false,
  });
  districtLayer.addTo(map);

  const markerLayer = L.layerGroup().addTo(map);
  function renderMarkers() {
    markerLayer.clearLayers();
    points.forEach((point) => {
      const count = Number(point.count) || 0;
      if (count < 1) return;
      const marker = L.marker([point.latitude, point.longitude], {
        icon: L.divIcon({
          className: "job-map-marker-wrap",
          html: `<span class="job-map-marker"><span>${count.toLocaleString()}</span></span>`,
          iconSize: [34, 34],
          iconAnchor: [17, 17],
        }),
        title: `${point.name}: ${count} matching ${count === 1 ? "job" : "jobs"}`,
      });
      marker.bindPopup(`<strong>${point.name}</strong><br>${count.toLocaleString()} matching ${count === 1 ? "job" : "jobs"}`);
      marker.addTo(markerLayer);
    });
  }
  renderMarkers();

  if (districtLayer.getBounds().isValid()) {
    map.fitBounds(districtLayer.getBounds(), { padding: [20, 20], maxZoom: 9 });
  }
  window.addEventListener("jobs:map-update", (event) => {
    districts = event.detail.districts;
    points = event.detail.points;
    holder.dataset.districts = JSON.stringify(districts);
    holder.dataset.points = JSON.stringify(points);
    districtByID = new Map(districts.map((district) => [district.id, district]));
    districtLayer.eachLayer((layer) => {
      districtLayer.resetStyle(layer);
    });
    renderMarkers();
  });
  window.setTimeout(() => map.invalidateSize(), 0);
  setStatus("");
}

function initTagSearch() {
  const form = document.querySelector("[data-tags]");
  if (!form) return;

  const input = form.querySelector("[data-tag-input]");
  const hidden = form.querySelector("[data-tags-value]");
  const selectedEl = form.querySelector("[data-selected-tags]");
  const suggestionsEl = form.querySelector("[data-tag-suggestions]");
  const yoeMin = form.querySelector("[data-yoe-min]");
  const yoeMax = form.querySelector("[data-yoe-max]");
  const yoeOutput = form.querySelector("[data-yoe-output]");
  const yoeTrack = form.querySelector("[data-yoe-track]");
  const tags = JSON.parse(form.dataset.tags || "[]");
  const selected = new Map();

  function selectedIDs() {
    return hidden.value.split(",").map((id) => id.trim()).filter(Boolean);
  }

  function syncSelectedFromHidden() {
    selected.clear();
    for (const id of selectedIDs()) {
      const tag = tags.find((item) => item.id === id);
      if (tag) selected.set(id, tag);
    }
  }

  function renderSelected() {
    selectedEl.innerHTML = "";
    for (const tag of selected.values()) {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "selected-tag";
      button.dataset.tagId = tag.id;
      button.innerHTML = `${tag.label} <span>&times;</span>`;
      selectedEl.appendChild(button);
    }
    hidden.value = Array.from(selected.keys()).join(",");
  }

  function matches(query, tag) {
    const aliases = Array.isArray(tag.aliases) ? tag.aliases.join(" ") : "";
    const text = `${tag.label} ${tag.category} ${tag.id} ${aliases}`.toLowerCase();
    return query.split(/\s+/).every((part) => text.includes(part));
  }

  function renderSuggestions() {
    const query = input.value.trim().toLowerCase();
    suggestionsEl.innerHTML = "";
    if (!query) {
      suggestionsEl.hidden = true;
      return;
    }

    const hits = tags.filter((tag) => !selected.has(tag.id) && matches(query, tag)).slice(0, 8);
    if (hits.length === 0) {
      suggestionsEl.hidden = true;
      return;
    }

    for (const tag of hits) {
      const button = document.createElement("button");
      button.type = "button";
      button.dataset.tagId = tag.id;
      button.innerHTML = `<strong>${tag.label}</strong><span>${tag.category}</span>`;
      suggestionsEl.appendChild(button);
    }
    suggestionsEl.hidden = false;
  }

  function addTag(id) {
    const tag = tags.find((item) => item.id === id);
    if (!tag) return;
    selected.set(tag.id, tag);
    input.value = "";
    renderSelected();
    renderSuggestions();
    requestLiveFilterUpdate(form);
  }

  syncSelectedFromHidden();
  renderSelected();

  input.addEventListener("input", () => {
    renderSuggestions();
  });

  input.addEventListener("keydown", (event) => {
    if (event.key !== "Enter") return;
    const first = suggestionsEl.querySelector("button[data-tag-id]");
    if (!first) return;
    event.preventDefault();
    addTag(first.dataset.tagId);
  });

  suggestionsEl.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-tag-id]");
    if (!button) return;
    addTag(button.dataset.tagId);
    input.focus();
  });

  selectedEl.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-tag-id]");
    if (!button) return;
    selected.delete(button.dataset.tagId);
    renderSelected();
    requestLiveFilterUpdate(form);
  });

  function renderYOE() {
    if (!yoeMin || !yoeMax || !yoeOutput || !yoeTrack) return;
    yoeOutput.textContent = `${yoeMin.value}–${yoeMax.value} years`;
    yoeTrack.style.setProperty("--range-start", `${Number(yoeMin.value) * 10}%`);
    yoeTrack.style.setProperty("--range-end", `${Number(yoeMax.value) * 10}%`);
  }

  if (yoeMin && yoeMax) {
    yoeMin.addEventListener("input", () => {
      if (Number(yoeMin.value) > Number(yoeMax.value)) yoeMax.value = yoeMin.value;
      renderYOE();
      requestLiveFilterUpdate(form);
    });
    yoeMax.addEventListener("input", () => {
      if (Number(yoeMax.value) < Number(yoeMin.value)) yoeMin.value = yoeMax.value;
      renderYOE();
      requestLiveFilterUpdate(form);
    });
    renderYOE();
  }
}

function initGovernmentPicker() {
  const picker = document.querySelector("[data-government-picker]");
  if (!picker) return;

  const queryInput = picker.querySelector("[data-government-query]");
  const summary = picker.querySelector("[data-government-summary]");
  const clearButton = picker.querySelector("[data-clear-governments]");
  const checkboxes = Array.from(picker.querySelectorAll('input[name="gov"]'));
  const options = Array.from(picker.querySelectorAll("[data-government-option]"));
  const groups = Array.from(picker.querySelectorAll("[data-government-group]"));

  function updateSummary() {
    const count = checkboxes.filter((checkbox) => checkbox.checked).length;
    summary.textContent = count === 0 ? "All governments" : `${count} selected`;
  }

  function updateRegion(group) {
    const regionCheckboxes = Array.from(group.querySelectorAll('input[name="gov"]'));
    const regionToggle = group.querySelector("[data-region-toggle]");
    const regionSummary = group.querySelector("[data-region-summary]");
    const selectedCount = regionCheckboxes.filter((checkbox) => checkbox.checked).length;
    regionToggle.checked = selectedCount === regionCheckboxes.length;
    regionToggle.indeterminate = selectedCount > 0 && selectedCount < regionCheckboxes.length;
    regionSummary.textContent = selectedCount === 0
      ? `${regionCheckboxes.length} governments`
      : `${selectedCount}/${regionCheckboxes.length} selected`;
  }

  function updateSelectionState() {
    updateSummary();
    groups.forEach(updateRegion);
  }

  function filterOptions() {
    const query = queryInput.value.trim().toLowerCase();
    options.forEach((option) => {
      option.hidden = query !== "" && !option.dataset.search.toLowerCase().includes(query);
    });
    groups.forEach((group) => {
      group.hidden = !group.querySelector("[data-government-option]:not([hidden])");
      if (query && !group.hidden) group.open = true;
    });
  }

  const form = picker.closest("form");
  checkboxes.forEach((checkbox) => checkbox.addEventListener("change", () => {
    updateSelectionState();
    requestLiveFilterUpdate(form);
  }));
  groups.forEach((group) => {
    const regionToggle = group.querySelector("[data-region-toggle]");
    regionToggle.addEventListener("change", () => {
      group.querySelectorAll('input[name="gov"]').forEach((checkbox) => {
        checkbox.checked = regionToggle.checked;
      });
      updateSelectionState();
      requestLiveFilterUpdate(form);
    });
  });
  queryInput.addEventListener("input", filterOptions);
  queryInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") event.preventDefault();
  });

  clearButton.addEventListener("click", () => {
    checkboxes.forEach((checkbox) => {
      checkbox.checked = false;
    });
    updateSelectionState();
    requestLiveFilterUpdate(form);
  });

  updateSelectionState();
}

function requestLiveFilterUpdate(form) {
  form.dispatchEvent(new CustomEvent("livefilterchange"));
}

function initLiveFilters() {
  const form = document.querySelector("form[data-tags]");
  if (!form) return;

  const status = form.querySelector("[data-live-status]");
  let timer;
  let controller;

  function setStatus(message) {
    if (status) status.textContent = message;
  }

  async function refresh() {
    window.clearTimeout(timer);
    if (controller) controller.abort();
    controller = new AbortController();

    const params = new URLSearchParams(new FormData(form));
    const url = `${form.action}?${params.toString()}`;
    const currentResults = document.querySelector(".results-column");
    currentResults.classList.add("is-updating");
    currentResults.setAttribute("aria-busy", "true");
    setStatus("Updating jobs...");

    try {
      const response = await fetch(url, {
        headers: { "X-Requested-With": "fetch" },
        signal: controller.signal,
      });
      if (!response.ok) throw new Error(`request returned ${response.status}`);

      const page = new DOMParser().parseFromString(await response.text(), "text/html");
      const nextResults = page.querySelector(".results-column");
      const nextMap = page.querySelector("[data-leaflet-map]");
      if (!nextResults || !nextMap) throw new Error("response did not contain filter results");

      currentResults.replaceWith(nextResults);
      const districts = JSON.parse(nextMap.dataset.districts || "[]");
      const points = JSON.parse(nextMap.dataset.points || "[]");
      window.dispatchEvent(new CustomEvent("jobs:map-update", { detail: { districts, points } }));
      window.history.replaceState({}, "", url);
      document.title = page.title;
      setStatus(`${nextResults.querySelectorAll(".job-card").length} jobs shown`);
    } catch (error) {
      if (error.name === "AbortError") return;
      console.error("Live filter update failed", error);
      currentResults.classList.remove("is-updating");
      currentResults.removeAttribute("aria-busy");
      setStatus("Could not update jobs. Try Search again.");
    }
  }

  function scheduleRefresh() {
    window.clearTimeout(timer);
    timer = window.setTimeout(refresh, 180);
  }

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    refresh();
  });
  form.addEventListener("livefilterchange", scheduleRefresh);
}

function initSignalHighlights() {
  const buttons = document.querySelectorAll("[data-signal-tag]");
  const lines = Array.from(document.querySelectorAll("[data-listing-line]"));
  if (buttons.length === 0 || lines.length === 0) return;

  function clearHighlights() {
    buttons.forEach((button) => button.classList.remove("is-active"));
    lines.forEach((line) => line.classList.remove("is-highlighted"));
  }

  function aliasesFor(button) {
    const highlightText = (button.dataset.highlightText || "").toLowerCase().trim();
    if (highlightText) return [highlightText];
    try {
      return JSON.parse(button.dataset.aliases || "[]")
        .map((alias) => alias.toLowerCase().trim())
        .filter(Boolean);
    } catch {
      return [];
    }
  }

  function lineMatches(line, aliases) {
    const text = line.textContent.toLowerCase();
    return aliases.some((alias) => text.includes(alias));
  }

  buttons.forEach((button) => {
    button.addEventListener("click", () => {
      const aliases = aliasesFor(button);
      clearHighlights();
      button.classList.add("is-active");

      const matches = lines.filter((line) => lineMatches(line, aliases));
      matches.forEach((line) => line.classList.add("is-highlighted"));

      if (matches[0]) {
        matches[0].scrollIntoView({ behavior: "smooth", block: "center" });
      }
    });
  });
}
