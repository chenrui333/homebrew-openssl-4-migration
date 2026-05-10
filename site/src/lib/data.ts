import snapshotJson from "../../../data/site_snapshot.json";
import type { Snapshot, SnapshotRow } from "./types";

export const snapshot = snapshotJson as Snapshot;

export const queueActions = [
  "Review and merge",
  "Retarget to staging",
  "Inspect CI",
  "Resolve draft blockers",
  "Track upstream blocker",
  "Open migration PR",
];

export const rowsByImpact = [...snapshot.rows].sort(compareRows);
export const registryRows = [...snapshot.rows].sort(compareRegistryRows);

export function compareRows(a: SnapshotRow, b: SnapshotRow) {
  if (a.impact_count !== b.impact_count) return b.impact_count - a.impact_count;
  return a.name.localeCompare(b.name);
}

export function compareRegistryRows(a: SnapshotRow, b: SnapshotRow) {
  if (a.flags.current_gate !== b.flags.current_gate) return a.flags.current_gate ? -1 : 1;
  if (isPending(a) !== isPending(b)) return isPending(a) ? -1 : 1;
  return compareRows(a, b);
}

export function currentGateRows(limit?: number) {
  const rows = rowsByImpact.filter((row) => row.flags.current_gate && isPending(row));
  return typeof limit === "number" ? rows.slice(0, limit) : rows;
}

export function rowsForAction(action: string, limit?: number) {
  const rows = rowsByImpact.filter((row) => row.next_action === action && isPending(row));
  return typeof limit === "number" ? rows.slice(0, limit) : rows;
}

export function actionCount(action: string) {
  return rowsForAction(action).length;
}

export function openCuratedBlockers() {
  return rowsByImpact.filter((row) => row.flags.upstream_blocked).sort(compareCurrentGateImpactRows);
}

export function upstreamCoverageGaps() {
  return rowsByImpact.filter((row) =>
    isPending(row) &&
    row.issues.some((issue) => issue.label === "search")
  );
}

export function reviewedNoIssueRows() {
  return rowsByImpact.filter((row) =>
    isPending(row) &&
    row.upstream.provider &&
    row.upstream.provider !== "other" &&
    row.issues.length === 0
  );
}

export function trackerFilterCounts() {
  return {
    "current-gate": snapshot.rows.filter((row) => row.flags.current_gate).length,
    "current-gate-blockers": snapshot.rows.filter((row) => row.flags.current_gate && isPending(row)).length,
    pending: snapshot.rows.filter((row) => isPending(row)).length,
    ready: snapshot.rows.filter((row) => row.flags.ready).length,
    draft: snapshot.rows.filter((row) => row.flags.draft).length,
    "ci-blocked": snapshot.rows.filter((row) => row.flags.ci_blocked).length,
    retarget: snapshot.rows.filter((row) => row.flags.base_mismatch).length,
    "missing-pr": snapshot.rows.filter((row) => row.flags.missing_pr).length,
    "upstream-blocked": snapshot.rows.filter((row) => row.flags.upstream_blocked).length,
    done: snapshot.rows.filter((row) => row.live_status === "DONE").length,
  };
}

export function gateCounts() {
  const counts = new Map<string, number>();
  snapshot.rows.forEach((row) => counts.set(row.group_label, (counts.get(row.group_label) || 0) + 1));
  return Array.from(counts.entries()).map(([label, count]) => ({ label, count }));
}

function compareCurrentGateImpactRows(a: SnapshotRow, b: SnapshotRow) {
  if (a.flags.current_gate !== b.flags.current_gate) return a.flags.current_gate ? -1 : 1;
  return compareRows(a, b);
}

function isPending(row: SnapshotRow) {
  return row.live_status === "PENDING";
}
