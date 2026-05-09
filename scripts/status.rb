#!/usr/bin/env ruby
# frozen_string_literal: true

require "json"
require "open3"
require "optparse"
require "pathname"
require "time"

DEPTH_LABELS = {
  0 => "Depth 0 (roots)",
  1 => "Depth 1",
  2 => "Depth 2",
  3 => "Depth 3",
  nil => "Leaves"
}.freeze

options = {
  homebrew_core: ENV.fetch("HOMEBREW_CORE", "/opt/homebrew/Library/Taps/homebrew/homebrew-core"),
  dep_tree: "data/dep_tree.json",
  output: "TRACKING.md"
}

OptionParser.new do |opts|
  opts.banner = "Usage: ruby scripts/status.rb [--homebrew-core=PATH] [--dep-tree=PATH] [--output=PATH]"
  opts.on("--homebrew-core=PATH", "Path to a homebrew-core checkout") { |path| options[:homebrew_core] = path }
  opts.on("--dep-tree=PATH", "Dependency inventory JSON") { |path| options[:dep_tree] = path }
  opts.on("--output=PATH", "Tracking Markdown path") { |path| options[:output] = path }
end.parse!

homebrew_core = Pathname.new(options[:homebrew_core]).expand_path
abort "homebrew-core checkout not found: #{homebrew_core}" unless homebrew_core.directory?

dep_tree_path = Pathname.new(options[:dep_tree])
abort "Dependency inventory not found: #{dep_tree_path}. Run make dep-tree first." unless dep_tree_path.file?

data = JSON.parse(dep_tree_path.read)
formulae = data.fetch("formulae")

live_status = lambda do |formula|
  path = homebrew_core/formula.fetch("path")
  unless path.file?
    fallback = homebrew_core/"Formula"/formula.fetch("name")[0]/"#{formula.fetch("name")}.rb"
    lib_fallback = homebrew_core/"Formula"/"lib"/"#{formula.fetch("name")}.rb"
    path = if fallback.file?
      fallback
    elsif lib_fallback.file?
      lib_fallback
    end
  end

  return ["REMOVED", nil] unless path&.file?

  contents = path.read
  status = if contents.match?(/^\s*depends_on\s+["']openssl@4["']/)
    "DONE"
  elsif contents.match?(/^\s*depends_on\s+["']openssl@3["']/)
    "PENDING"
  else
    "UNKNOWN"
  end

  [status, path]
end

pr_by_formula = {}
pr_warning = nil
stdout, stderr, status = Open3.capture3(
  "gh", "pr", "list",
  "--repo", "Homebrew/homebrew-core",
  "--search", "openssl@4",
  "--state", "open",
  "--json", "number,title,state",
  "--limit", "1000"
)

if status.success?
  JSON.parse(stdout).each do |pr|
    title = pr.fetch("title")
    next unless (match = title.match(/\A(.+?): use openssl@4\z/))

    pr_by_formula[match[1]] = pr
  end
else
  pr_warning = stderr.strip.empty? ? "gh pr list failed" : stderr.strip
end

rows = formulae.map do |formula|
  status_text, path = live_status.call(formula)
  formula.merge(
    "live_status" => status_text,
    "live_path" => path&.relative_path_from(homebrew_core)&.to_s,
    "open_pr" => pr_by_formula[formula.fetch("name")]
  )
end

pending_count = rows.count { |row| row.fetch("live_status") == "PENDING" }
done_count = rows.count { |row| row.fetch("live_status") == "DONE" }
total_count = pending_count + done_count
done_percent = total_count.zero? ? 0.0 : ((done_count.to_f / total_count) * 100)

date = Time.now.strftime("%Y-%m-%d")
lines = []
lines << "OpenSSL 4 Migration Status (#{date})"
lines << "========================================"
lines << "Total pending:  #{pending_count}"
lines << format("Total done:     %<done>d (%<percent>.1f%%)", done: done_count, percent: done_percent)
lines << ""
lines << "Tracking issue: Homebrew/homebrew-core#278366"
lines << ""
lines << "Warning: #{pr_warning}" if pr_warning
lines << "" if pr_warning

[0, 1, 2, 3, nil].each do |depth|
  group = rows.select { |row| row["depth"] == depth }.sort_by { |row| row.fetch("name") }
  next if group.empty?

  group_done = group.count { |row| row.fetch("live_status") == "DONE" }
  lines << "#{DEPTH_LABELS.fetch(depth)}   [#{group_done}/#{group.length} done]"

  group.each do |row|
    pr = row["open_pr"]
    suffix = pr ? "   [PR ##{pr.fetch("number")} open]" : ""
    lines << format("  %-24s %-8s%s", row.fetch("name"), row.fetch("live_status"), suffix).rstrip
  end

  lines << ""
end

output = lines.join("\n").rstrip + "\n"
puts output
Pathname.new(options[:output]).write(output)
