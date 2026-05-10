export interface IssueLink {
  url: string;
  label: string;
  title?: string;
  state?: string;
  status?: string;
}

export interface SnapshotSummary {
  staged_formulae: number;
  done: number;
  pending: number;
  open_staging_prs: number;
  current_gate: string;
  current_gate_pending: number;
  upstream_blockers: number;
  base_mismatches: number;
}

export interface PullRequest {
  number: number;
  url: string;
  base: string;
  is_draft: boolean;
  merge_state: string;
  updated_at: string;
}

export interface Upstream {
  provider: string;
  repo: string;
  url: string;
}

export interface RowFlags {
  current_gate: boolean;
  ready: boolean;
  draft: boolean;
  ci_blocked: boolean;
  base_mismatch: boolean;
  missing_pr: boolean;
  upstream_blocked: boolean;
}

export interface SnapshotRow {
  name: string;
  live_status: string;
  depth: number | null;
  group_label: string;
  impact_count: number;
  target_branch: string;
  pr: PullRequest;
  readiness: string[];
  next_action: string;
  upstream: Upstream;
  issues: IssueLink[];
  flags: RowFlags;
}

export interface Snapshot {
  generated_at: string;
  repository: string;
  tracking_issue: string;
  target_branch: string;
  scope: "staging";
  summary: SnapshotSummary;
  rows: SnapshotRow[];
}
