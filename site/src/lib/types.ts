export interface IssueLink {
  url: string;
  label: string;
  title?: string;
  state?: string;
  status?: string;
}

export interface GateSnapshot {
  label: string;
  depth: number | null;
  target_branch: string;
  done: number;
  total: number;
  pending: number;
}

export interface SnapshotRow {
  name: string;
  live_status: string;
  target_branch: string;
  depth: number | null;
  staging_reason?: string;
  impact_count: number;
  open_pr_number?: number;
  open_pr_url?: string;
  open_pr_base?: string;
  readiness: string[];
  next_action: string;
  upstream_provider?: string;
  upstream_repo?: string;
  upstream_url?: string;
  issue_links: IssueLink[];
  is_current_gate: boolean;
  is_upstream_blocked: boolean;
  is_base_mismatch: boolean;
  is_ci_blocked: boolean;
  is_draft: boolean;
  is_missing_pr: boolean;
  is_ready: boolean;
}

export interface Snapshot {
  generated_at: string;
  total_formulae: number;
  done: number;
  pending: number;
  open_prs: number;
  current_gate: GateSnapshot;
  next_gate?: GateSnapshot;
  upstream_gap_count: number;
  base_mismatch_count: number;
  rows: SnapshotRow[];
}
