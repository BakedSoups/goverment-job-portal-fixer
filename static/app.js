(function () {
  initTagSearch();
  initSignalHighlights();
})();

function initTagSearch() {
  const form = document.querySelector("[data-tags]");
  if (!form) return;

  const input = form.querySelector("[data-tag-input]");
  const hidden = form.querySelector("[data-tags-value]");
  const selectedEl = form.querySelector("[data-selected-tags]");
  const suggestionsEl = form.querySelector("[data-tag-suggestions]");
  const yoeSlider = form.querySelector("[data-yoe-slider]");
  const yoeOutput = form.querySelector("[data-yoe-output]");
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
    const text = `${tag.label} ${tag.category} ${tag.id}`.toLowerCase();
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
  }

  syncSelectedFromHidden();
  renderSelected();

  input.addEventListener("input", renderSuggestions);

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
  });

  function renderYOE() {
    if (!yoeSlider || !yoeOutput) return;
    yoeOutput.textContent = `${yoeSlider.value} ${yoeSlider.value === "1" ? "year" : "years"}`;
  }

  if (yoeSlider) {
    yoeSlider.addEventListener("input", renderYOE);
    renderYOE();
  }
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
