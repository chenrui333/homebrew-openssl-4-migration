(() => {
  const normalize = (value) => (value || "").toLowerCase();

  const rowText = (row) => [
    row.dataset.formula,
    row.dataset.status,
    row.dataset.readiness,
    row.dataset.action,
    row.dataset.target,
    row.textContent,
  ].map(normalize).join(" ");

  const matchesQuickFilter = (row, filter) => {
    const readiness = " " + normalize(row.dataset.readiness) + " ";
    const action = normalize(row.dataset.action);
    const status = normalize(row.dataset.status);

    switch (filter) {
      case "pending":
        return status === "pending";
      case "ready":
        return action === "ready";
      case "draft":
        return readiness.includes(" draft ");
      case "ci-blocked":
        return readiness.includes(" checks-blocked ") || readiness.includes(" no-checks ") || action === "ci-blocked";
      case "base-mismatch":
        return readiness.includes(" base-mismatch ");
      case "missing-pr":
        return readiness.includes(" missing-pr ") || action === "open-pr";
      case "upstream-blocked":
        return row.dataset.upstreamBlocked === "true";
      case "all":
      default:
        return true;
    }
  };

  const updateDetails = (scope) => {
    scope.querySelectorAll("details.tracker-group").forEach((details) => {
      const rows = Array.from(details.querySelectorAll("tr[data-status]"));
      if (rows.length === 0) return;
      const visible = rows.some((row) => !row.hidden);
      details.classList.toggle("is-filter-empty", !visible);
      if (visible) details.open = true;
    });
  };

  const initControls = (controls) => {
    if (controls.dataset.ready === "true") return;
    controls.dataset.ready = "true";

    const scope = controls.parentElement || document;
    const input = controls.querySelector("[data-filter-text]");
    const buttons = Array.from(controls.querySelectorAll("[data-quick-filter]"));
    const count = controls.querySelector("[data-filter-count]");
    let active = "all";

    const apply = () => {
      const query = normalize(input ? input.value : "");
      const rows = Array.from(scope.querySelectorAll("tr[data-status]"));
      let visible = 0;

      rows.forEach((row) => {
        const show = matchesQuickFilter(row, active) && rowText(row).includes(query);
        row.hidden = !show;
        if (show) visible += 1;
      });

      buttons.forEach((button) => {
        button.setAttribute("aria-pressed", button.dataset.quickFilter === active ? "true" : "false");
      });
      if (count) count.textContent = visible + " of " + rows.length + " rows shown";
      updateDetails(scope);
    };

    buttons.forEach((button) => {
      button.addEventListener("click", () => {
        active = button.dataset.quickFilter || "all";
        apply();
      });
    });
    if (input) input.addEventListener("input", apply);
    apply();
  };

  const sortValue = (row, key) => {
    if (key === "impact") return Number(row.dataset.impact || 0);
    if (key === "formula") return normalize(row.dataset.formula);
    if (key === "status") return normalize(row.dataset.status);
    if (key === "action") return normalize(row.dataset.action);
    return "";
  };

  const initTableSort = (table) => {
    if (table.dataset.sortReady === "true") return;
    table.dataset.sortReady = "true";

    table.querySelectorAll("[data-sort]").forEach((button) => {
      button.addEventListener("click", () => {
        const key = button.dataset.sort;
        const tbody = table.tBodies[0];
        if (!tbody) return;

        const previousKey = table.dataset.sortKey;
        const previousDirection = table.dataset.sortDirection || "desc";
        const direction = previousKey === key && previousDirection === "desc" ? "asc" : "desc";
        const rows = Array.from(tbody.querySelectorAll("tr[data-status]"));

        rows.sort((a, b) => {
          const av = sortValue(a, key);
          const bv = sortValue(b, key);
          let result = 0;
          if (typeof av === "number" && typeof bv === "number") {
            result = av - bv;
          } else {
            result = String(av).localeCompare(String(bv));
          }
          if (result === 0) result = normalize(a.dataset.formula).localeCompare(normalize(b.dataset.formula));
          return direction === "asc" ? result : -result;
        });

        rows.forEach((row) => tbody.appendChild(row));
        table.dataset.sortKey = key || "";
        table.dataset.sortDirection = direction;
      });
    });
  };

  const init = () => {
    document.querySelectorAll("[data-tracker-controls]").forEach(initControls);
    document.querySelectorAll("table[data-tracker-table]").forEach(initTableSort);
  };

  if (typeof document$ !== "undefined") {
    document$.subscribe(init);
  } else if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
