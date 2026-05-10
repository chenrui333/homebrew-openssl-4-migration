import type { SnapshotRow } from "./types";

export const basePath = "/homebrew-openssl-4-migration/";

export function pagePath(path = "") {
  return (basePath + path).replace(/\/{2,}/g, "/");
}

export function formatDate(value: string) {
  if (!value) return "unknown";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function pct(done: number, total: number) {
  if (total === 0) return "0.0%";
  return ((done / total) * 100).toFixed(1) + "%";
}

export function slug(value: string) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

export function formulaAnchor(row: SnapshotRow | string) {
  return slug(typeof row === "string" ? row : row.name);
}

export function sectionAnchor(label: string) {
  return slug(label);
}

export function trackerHref(row: SnapshotRow) {
  return pagePath("tracker/#" + formulaAnchor(row));
}

export function prLabel(row: SnapshotRow) {
  return row.pr.number > 0 ? "#" + row.pr.number : "No PR";
}

export function upstreamLabel(row: SnapshotRow) {
  if (row.upstream.repo) return row.upstream.provider + ":" + row.upstream.repo;
  return row.upstream.provider || "other";
}

export function depthLabel(row: SnapshotRow) {
  return row.depth === null || row.depth === undefined ? "leaf" : "depth " + row.depth;
}
