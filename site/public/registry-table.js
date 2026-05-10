(function () {
  function registryTables() {
    return Array.from(document.querySelectorAll("[data-registry-table]"));
  }

  function registryRows() {
    return registryTables().flatMap((table) => Array.from(table.querySelectorAll("[data-row]")));
  }

  function matchesFilter(row, filter) {
    if (filter === "all") return true;
    if (filter === "current-gate") return row.dataset.currentGate === "true";
    if (filter === "pending") return row.dataset.status === "pending";
    if (filter === "done") return row.dataset.status === "done";
    if (filter === "ready") return row.dataset.ready === "true";
    if (filter === "draft") return row.dataset.draft === "true";
    if (filter === "ci-blocked") return row.dataset.ciBlocked === "true";
    if (filter === "retarget") return row.dataset.baseMismatch === "true";
    if (filter === "missing-pr") return row.dataset.missingPr === "true";
    if (filter === "upstream-blocked") return row.dataset.upstreamBlocked === "true";
    if (filter === "has-pr") return row.dataset.pr === "true";
    if (filter === "no-pr") return row.dataset.pr !== "true";
    if (filter === "has-upstream-issue") return row.dataset.hasIssue === "true";
    if (filter === "coverage-gap") return row.dataset.coverageGap === "true";
    if (filter.startsWith("gate:")) return row.dataset.gateFilter === filter.slice(5);
    return true;
  }

  function setFilterState(active) {
    document.querySelectorAll("[data-filter]").forEach((button) => {
      button.setAttribute("aria-pressed", button.dataset.filter === active ? "true" : "false");
    });
  }

  function applyFilters(active) {
    const query = String(document.querySelector("[data-registry-search]")?.value || "").toLowerCase();
    const rows = registryRows();
    let visible = 0;
    rows.forEach((row) => {
      const shown = matchesFilter(row, active) && String(row.dataset.filterText || "").includes(query);
      row.hidden = !shown;
      if (shown) visible += 1;
    });
    setFilterState(active);
    document.querySelectorAll("[data-result-count]").forEach((count) => {
      count.textContent = visible + " of " + rows.length + " rows shown";
    });
  }

  function valueFor(row, key) {
    if (key === "impact") return Number(row.dataset.impact || 0);
    if (key === "action") return String(row.dataset.action || "");
    if (key === "gate") return String(row.dataset.gate || "");
    if (key === "status") return String(row.dataset.status || "");
    return String(row.dataset.name || "");
  }

  function sortTable(table, sortKey, sortDirection) {
    const tbody = table.querySelector("tbody");
    if (!tbody) return;
    const rows = Array.from(table.querySelectorAll("[data-row]"));
    const sorted = rows.sort((a, b) => {
      if (sortKey === "impact") {
        const currentDelta = Number(b.dataset.currentGate === "true") - Number(a.dataset.currentGate === "true");
        if (currentDelta !== 0) return currentDelta;
        const pendingDelta = Number(b.dataset.status === "pending") - Number(a.dataset.status === "pending");
        if (pendingDelta !== 0) return pendingDelta;
      }
      const av = valueFor(a, sortKey);
      const bv = valueFor(b, sortKey);
      let result = 0;
      if (typeof av === "number" && typeof bv === "number") {
        result = av - bv;
      } else {
        result = String(av).localeCompare(String(bv));
      }
      if (result !== 0) return sortDirection === "asc" ? result : -result;
      return String(a.dataset.name || "").localeCompare(String(b.dataset.name || ""));
    });
    sorted.forEach((row) => tbody.appendChild(row));
  }

  function updateSortButtons(table, sortKey, sortDirection) {
    table.querySelectorAll("[data-sort]").forEach((button) => {
      const activeButton = button.dataset.sort === sortKey;
      button.toggleAttribute("data-sort-active", activeButton);
      if (activeButton) {
        button.dataset.sortDirection = sortDirection;
      } else {
        delete button.dataset.sortDirection;
      }
      button.setAttribute("aria-label", (button.textContent || "Column") + " sorted " + (activeButton ? sortDirection : "none"));
      const th = button.closest("th");
      if (th) th.setAttribute("aria-sort", activeButton ? (sortDirection === "asc" ? "ascending" : "descending") : "none");
    });
  }

  function initTable(table) {
    if (table.dataset.registryReady === "true") return;
    table.dataset.registryReady = "true";
    let sortKey = "impact";
    let sortDirection = "desc";
    table.querySelectorAll("[data-sort]").forEach((button) => {
      button.addEventListener("click", () => {
        const next = button.dataset.sort || "impact";
        if (next === sortKey) {
          sortDirection = sortDirection === "asc" ? "desc" : "asc";
        } else {
          sortKey = next;
          sortDirection = next === "impact" ? "desc" : "asc";
        }
        sortTable(table, sortKey, sortDirection);
        updateSortButtons(table, sortKey, sortDirection);
        applyFilters(window.__registryActiveFilter || "all");
      });
    });
    sortTable(table, sortKey, sortDirection);
    updateSortButtons(table, sortKey, sortDirection);
  }

  function init() {
    window.__registryActiveFilter = window.__registryActiveFilter || "all";
    registryTables().forEach(initTable);
    document.querySelectorAll("[data-filter]").forEach((button) => {
      button.addEventListener("click", () => {
        window.__registryActiveFilter = button.dataset.filter || "all";
        applyFilters(window.__registryActiveFilter);
      });
    });
    document.querySelectorAll("[data-registry-search]").forEach((input) => {
      input.addEventListener("input", () => applyFilters(window.__registryActiveFilter || "all"));
    });
    applyFilters(window.__registryActiveFilter || "all");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
