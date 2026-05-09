#!/usr/bin/env ruby
# frozen_string_literal: true

require "json"
require "open3"
require "optparse"
require "pathname"
require "shellwords"
require "tempfile"

STAGING_BRANCH = "openssl-4-migration-staging"
TRACKING_ISSUE = "Homebrew/homebrew-core#278366"

options = {
  homebrew_core: ENV.fetch("HOMEBREW_CORE", "/opt/homebrew/Library/Taps/homebrew/homebrew-core"),
  dep_tree: "data/dep_tree.json",
  dry_run: false,
  no_pr: false,
  push_remote: ENV["HOMEBREW_CORE_PUSH_REMOTE"]
}

parser = OptionParser.new do |opts|
  opts.banner = "Usage: ruby scripts/migrate.rb <formula-name> [--dry-run] [--no-pr] [--homebrew-core=PATH]"
  opts.on("--homebrew-core=PATH", "Path to a homebrew-core checkout") { |path| options[:homebrew_core] = path }
  opts.on("--dep-tree=PATH", "Dependency inventory JSON") { |path| options[:dep_tree] = path }
  opts.on("--dry-run", "Print the planned diff without changing homebrew-core") { options[:dry_run] = true }
  opts.on("--no-pr", "Create the local branch and commit, but skip push and PR creation") { options[:no_pr] = true }
  opts.on("--push-remote=REMOTE", "Fork remote to push to before creating the PR") { |remote| options[:push_remote] = remote }
end
parser.parse!

formula_name = ARGV.shift
abort parser.to_s unless formula_name
abort "Unexpected arguments: #{ARGV.join(" ")}" unless ARGV.empty?

homebrew_core = Pathname.new(options[:homebrew_core]).expand_path
abort "homebrew-core checkout not found: #{homebrew_core}" unless homebrew_core.directory?

dep_tree_path = Pathname.new(options[:dep_tree])
abort "Dependency inventory not found: #{dep_tree_path}. Run make dep-tree first." unless dep_tree_path.file?

data = JSON.parse(dep_tree_path.read)
formula_entry = data.fetch("formulae").find { |entry| entry.fetch("name") == formula_name }
abort "#{formula_name} is not present in #{dep_tree_path}" unless formula_entry

depth = formula_entry["depth"]
base_branch = depth.nil? ? "main" : STAGING_BRANCH
branch_name = "rchen.openssl4.#{formula_name}"
title = "#{formula_name}: use openssl@4"

CommandResult = Struct.new(:stdout, :stderr, :status, keyword_init: true) do
  def success?
    status.success?
  end
end

def run_command(*args, chdir:, allow_failure: false)
  stdout, stderr, status = Open3.capture3(*args, chdir: chdir.to_s)
  result = CommandResult.new(stdout: stdout, stderr: stderr, status: status)
  return result if allow_failure || result.success?

  warn stdout unless stdout.empty?
  warn stderr unless stderr.empty?
  abort "Command failed: #{args.shelljoin}"
end

def locate_formula(homebrew_core, formula_name, formula_entry = nil)
  candidates = []
  candidates << homebrew_core/formula_entry.fetch("path") if formula_entry&.fetch("path", nil)
  candidates << homebrew_core/"Formula"/"lib"/"#{formula_name}.rb" if formula_name.start_with?("lib")
  candidates << homebrew_core/"Formula"/formula_name[0]/"#{formula_name}.rb"
  candidates << homebrew_core/"Formula"/"#{formula_name}.rb"

  candidates.uniq.find(&:file?) || begin
    matches = (homebrew_core/"Formula").glob("**/#{formula_name}.rb")
    matches.first
  end
end

def bump_revision(contents)
  lines = contents.lines
  revision_index = lines.index { |line| line.match?(/^  revision \d+\s*$/) }

  if revision_index
    current = lines[revision_index].match(/^  revision (\d+)\s*$/)[1].to_i
    lines[revision_index] = "  revision #{current + 1}\n"
    return lines.join
  end

  bottle_start = lines.index { |line| line.match?(/^  bottle do\s*$/) }
  if bottle_start
    bottle_end = ((bottle_start + 1)...lines.length).find { |index| lines[index].match?(/^  end\s*$/) }
    if bottle_end
      insert_at = bottle_end + 1
      insert_at += 1 if lines[insert_at]&.match?(/^\s*$/)
      lines.insert(insert_at, "  revision 1\n", "\n")
      return lines.join
    end
  end

  insert_at = lines.index { |line| line.match?(/^  (depends_on|uses_from_macos|on_|def install)\b/) } || 1
  lines.insert(insert_at, "  revision 1\n", "\n")
  lines.join
end

def rust_formula?(contents)
  contents.match?(/^\s*depends_on\s+["']rust["'].*:build/) ||
    contents.match?(/\bsystem\s+["']cargo["']/) ||
    contents.match?(/\bstd_cargo_args\b/) ||
    contents.match?(/\bcargo\s+install\b/)
end

def add_rust_openssl_env(contents)
  return contents if contents.include?("OPENSSL_DIR")

  lines = contents.lines
  install_index = lines.index { |line| line.match?(/^  def install\s*$/) }
  return contents unless install_index

  env_lines = [
    "    ENV[\"OPENSSL_DIR\"] = Formula[\"openssl@4\"].opt_prefix\n",
    "    ENV[\"OPENSSL_LIB_DIR\"] = Formula[\"openssl@4\"].opt_lib\n",
    "    ENV[\"OPENSSL_INCLUDE_DIR\"] = Formula[\"openssl@4\"].opt_include\n",
    "    ENV.prepend_path \"PKG_CONFIG_PATH\", Formula[\"openssl@4\"].opt_lib/\"pkgconfig\"\n",
    "\n"
  ]

  lines.insert(install_index + 1, *env_lines)
  lines.join
end

def migrate_contents(contents)
  return :already_migrated if contents.match?(/^\s*depends_on\s+["']openssl@4["']/)
  abort "Formula does not depend on openssl@3" unless contents.match?(/^\s*depends_on\s+["']openssl@3["']/)

  migrated = contents.gsub(/depends_on "openssl@3"/, "depends_on \"openssl@4\"")
  migrated = bump_revision(migrated)
  migrated = add_rust_openssl_env(migrated) if rust_formula?(migrated)
  migrated
end

def print_diff(original, migrated, formula_name)
  Tempfile.create(["#{formula_name}-before", ".rb"]) do |before|
    Tempfile.create(["#{formula_name}-after", ".rb"]) do |after|
      before.write(original)
      before.flush
      after.write(migrated)
      after.flush
      stdout, = Open3.capture2("git", "diff", "--no-index", "--", before.path, after.path)
      puts stdout
    end
  end
end

def clean_worktree!(homebrew_core)
  status = run_command("git", "status", "--porcelain", chdir: homebrew_core).stdout
  return if status.empty?

  abort "homebrew-core working tree is not clean:\n#{status}"
end

def remote_url(homebrew_core, remote)
  run_command("git", "remote", "get-url", "--push", remote, chdir: homebrew_core, allow_failure: true).stdout.strip
end

def gh_login
  stdout, _stderr, status = Open3.capture3("gh", "api", "user", "--jq", ".login")
  status.success? ? stdout.strip : nil
end

def detect_push_remote(homebrew_core)
  login = gh_login
  remotes = run_command("git", "remote", chdir: homebrew_core).stdout.lines.map(&:strip)

  if login && !login.empty?
    matching = remotes.find do |remote|
      url = remote_url(homebrew_core, remote)
      url.match?(%r{[:/]#{Regexp.escape(login)}/homebrew-core(?:\.git)?\z})
    end
    return matching if matching
  end

  remotes.include?("origin") ? "origin" : remotes.first
end

def owner_from_remote_url(url)
  return Regexp.last_match(1) if url.match(%r{github\.com[:/]([^/]+)/homebrew-core(?:\.git)?\z})

  nil
end

formula_path = locate_formula(homebrew_core, formula_name, formula_entry)
abort "Formula file not found for #{formula_name}" unless formula_path&.file?

original = formula_path.read
migrated = migrate_contents(original)

if migrated == :already_migrated
  puts "#{formula_name} already depends on openssl@4; nothing to do."
  exit 0
end

puts "Formula: #{formula_name}"
puts "Path: #{formula_path.relative_path_from(homebrew_core)}"
puts "Base: #{base_branch}"
puts "Branch: #{branch_name}"

if options[:dry_run]
  print_diff(original, migrated, formula_name)
  exit 0
end

clean_worktree!(homebrew_core)
run_command("git", "fetch", "origin", base_branch, chdir: homebrew_core)

current_branch = run_command("git", "branch", "--show-current", chdir: homebrew_core).stdout.strip
branch_exists = run_command("git", "rev-parse", "--verify", branch_name, chdir: homebrew_core, allow_failure: true).success?

if current_branch != branch_name
  if branch_exists
    run_command("git", "switch", branch_name, chdir: homebrew_core)
  else
    run_command("git", "switch", "-c", branch_name, "origin/#{base_branch}", chdir: homebrew_core)
  end
end

formula_path = locate_formula(homebrew_core, formula_name, formula_entry)
abort "Formula file not found for #{formula_name} after switching branches" unless formula_path&.file?

original = formula_path.read
migrated = migrate_contents(original)
if migrated == :already_migrated
  puts "#{formula_name} already depends on openssl@4 on #{branch_name}; nothing to do."
  exit 0
end

formula_path.write(migrated)
diff = run_command("git", "diff", "--", formula_path.relative_path_from(homebrew_core).to_s, chdir: homebrew_core).stdout
if diff.empty?
  puts "No changes produced for #{formula_name}."
  exit 0
end

puts diff
run_command("git", "add", formula_path.relative_path_from(homebrew_core).to_s, chdir: homebrew_core)
run_command("git", "commit", "-s", "-m", title, chdir: homebrew_core)

if options[:no_pr]
  puts "Created local commit on #{branch_name}; skipping push and PR because --no-pr was set."
  exit 0
end

push_remote = options[:push_remote] || detect_push_remote(homebrew_core)
abort "Could not determine a push remote" unless push_remote

run_command("git", "push", "-u", push_remote, branch_name, chdir: homebrew_core)

owner = owner_from_remote_url(remote_url(homebrew_core, push_remote))
abort "Could not determine GitHub owner for remote #{push_remote}" unless owner

body = <<~BODY
  Migrates #{formula_name} from openssl@3 to openssl@4 as part of the staging migration.

  References:
  - #{TRACKING_ISSUE}
BODY

Tempfile.create(["openssl4-pr-body", ".md"]) do |body_file|
  body_file.write(body)
  body_file.flush

  pr_stdout = run_command(
    "gh", "pr", "create",
    "--repo", "Homebrew/homebrew-core",
    "--head", "#{owner}:#{branch_name}",
    "--base", base_branch,
    "--title", title,
    "--body-file", body_file.path,
    chdir: homebrew_core
  ).stdout.strip

  puts pr_stdout
  labels = ["openssl-4-migration"]
  labels += ["staging-branch-pr", "CI-skip-recursive-dependents"] if base_branch == STAGING_BRANCH
  label_result = run_command(
    "gh", "pr", "edit", pr_stdout,
    "--repo", "Homebrew/homebrew-core",
    *labels.flat_map { |label| ["--add-label", label] },
    chdir: homebrew_core,
    allow_failure: true
  )

  unless label_result.success?
    warn "Warning: PR created but labels could not be applied: #{label_result.stderr.strip}"
  end
end
