import snapshotJson from "../../../data/site_snapshot.json";
import type { Snapshot, SnapshotRow } from "./types";

export const snapshot = snapshotJson as Snapshot;

export const rowsByImpact = [...snapshot.rows].sort(compareRows);

export function compareRows(a: SnapshotRow, b: SnapshotRow) {
  if (a.impact_count !== b.impact_count) return b.impact_count - a.impact_count;
  return a.name.localeCompare(b.name);
}

export function currentGateRows(limit?: number) {
  const rows = rowsByImpact.filter((row) => row.is_current_gate && row.live_status === "PENDING");
  return typeof limit === "number" ? rows.slice(0, limit) : rows;
}

export function rowsForAction(action: string, limit?: number) {
  const rows = rowsByImpact.filter((row) => row.next_action === action && row.live_status === "PENDING");
  return typeof limit === "number" ? rows.slice(0, limit) : rows;
}

export function openCuratedBlockers() {
  return rowsByImpact.filter((row) => row.is_upstream_blocked);
}

export function upstreamCoverageGaps() {
  return rowsByImpact.filter((row) =>
    row.live_status === "PENDING" &&
    row.target_branch === "openssl-4-migration-staging" &&
    row.issue_links.some((issue) => issue.label === "search")
  );
}

export function reviewedNoIssueRows() {
  return rowsByImpact.filter((row) =>
    row.live_status === "PENDING" &&
    row.target_branch === "openssl-4-migration-staging" &&
    row.upstream_provider &&
    row.upstream_provider !== "other" &&
    row.issue_links.length === 0
  );
}
