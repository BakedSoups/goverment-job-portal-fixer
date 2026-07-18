import posthog from "https://cdn.jsdelivr.net/npm/posthog-js@1.404.1/+esm";

posthog.init("phc_toh6VupKc35Vb3kfg6aFMi5F3sLPuLtNRjarT3hjLrnG", {
  api_host: "https://us.i.posthog.com",
  defaults: "2026-05-30",
  autocapture: false,
  disable_session_recording: true,
  capture_pageview: true,
  capture_pageleave: true,
});

document.addEventListener("click", (event) => {
  const target = event.target.closest("[data-analytics-event]");
  if (!target) return;

  posthog.capture(target.dataset.analyticsEvent, {
    source: target.dataset.analyticsSource || undefined,
    region: target.dataset.analyticsRegion || undefined,
  });
});

document.addEventListener("submit", (event) => {
  const form = event.target.closest("form[data-tags]");
  if (!form) return;

  posthog.capture("job_search_submitted", {
    selected_skill_count: form.querySelectorAll("[data-selected-tags] [data-tag-id]").length,
    selected_government_count: form.querySelectorAll('input[name="gov"]:checked').length,
  });
});
