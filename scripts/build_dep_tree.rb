#!/usr/bin/env ruby
# frozen_string_literal: true

require "json"
require "open3"
require "optparse"
require "pathname"
require "set"
require "time"

TRACKING_ISSUE = "Homebrew/homebrew-core#278366"
STAGING_BRANCH = "openssl-4-migration-staging"

STAGED_DEPTHS = {
  0 => %w[
    cmake apr-util asio dotnet erlang freetds grpc hiredis krb5 libevent
    libfido2 librdkafka libssh libssh2 mariadb-connector-c openldap opusfile
    python@3.11 python@3.12 python@3.13 python@3.14 srt tcl-tk tcl-tk@8 wget
  ],
  1 => %w[
    apache-arrow bind curl ffmpeg folly httpd libpq node postgresql@17
    postgresql@18 pulseaudio qtbase rust systemd unbound
  ],
  2 => %w[cargo-c cryptography gdal php ruby],
  3 => %w[gstreamer]
}.freeze

options = {
  homebrew_core: ENV.fetch("HOMEBREW_CORE", "/opt/homebrew/Library/Taps/homebrew/homebrew-core"),
  output: "data/dep_tree.json"
}

OptionParser.new do |opts|
  opts.banner = "Usage: ruby scripts/build_dep_tree.rb [--homebrew-core=PATH] [--output=PATH]"
  opts.on("--homebrew-core=PATH", "Path to a homebrew-core checkout") { |path| options[:homebrew_core] = path }
  opts.on("--output=PATH", "Output JSON path") { |path| options[:output] = path }
end.parse!

homebrew_core = Pathname.new(options[:homebrew_core]).expand_path
formula_root = homebrew_core/"Formula"
abort "Formula directory not found: #{formula_root}" unless formula_root.directory?

staged_depth_by_formula = STAGED_DEPTHS.each_with_object({}) do |(depth, names), memo|
  names.each { |name| memo[name] = depth }
end

formulae = []

formula_root.glob("**/*.rb").sort.each do |path|
  name = path.basename(".rb").to_s
  contents = path.read

  openssl_dependency = if contents.match?(/^\s*depends_on\s+["']openssl@4["']/)
    "openssl@4"
  elsif contents.match?(/^\s*depends_on\s+["']openssl@3["']/)
    "openssl@3"
  end

  dependencies = contents.each_line.filter_map do |line|
    match = line.match(/^\s*depends_on\s+["']([^"']+)["'](.*)$/)
    next unless match

    dep_name = match[1]
    qualifier = match[2]
    test_only = qualifier.match?(/=>\s*:test\b/) && !qualifier.match?(/=>\s*\[[^\]]*:build/)
    next if test_only

    dep_name
  end.uniq.sort

  next unless openssl_dependency

  formulae << {
    "name" => name,
    "path" => path.relative_path_from(homebrew_core).to_s,
    "openssl_dependency" => openssl_dependency,
    "depth" => staged_depth_by_formula[name],
    "dependencies" => dependencies
  }
end

tracked_names = formulae.map { |formula| formula.fetch("name") }.to_set

formulae.each do |formula|
  deps = formula.fetch("dependencies").select { |dependency| tracked_names.include?(dependency) }
  formula["openssl_formula_dependencies"] = deps
end

dependents = Hash.new { |hash, key| hash[key] = [] }
formulae.each do |formula|
  formula.fetch("openssl_formula_dependencies").each do |dependency|
    dependents[dependency] << formula.fetch("name")
  end
end

formulae.each do |formula|
  formula["openssl_formula_dependents"] = dependents[formula.fetch("name")].sort
end

missing_staged = staged_depth_by_formula.keys - formulae.map { |formula| formula.fetch("name") }
warn "Warning: staged formulae missing from OpenSSL inventory: #{missing_staged.sort.join(", ")}" unless missing_staged.empty?

git_head_stdout, = Open3.capture2("git", "rev-parse", "HEAD", chdir: homebrew_core.to_s)
git_head = git_head_stdout.strip

data = {
  "generated_at" => Time.now.iso8601,
  "repository" => "Homebrew/homebrew-core",
  "git_head" => git_head.empty? ? nil : git_head,
  "tracking_issue" => TRACKING_ISSUE,
  "staging_branch" => STAGING_BRANCH,
  "staged_depths" => STAGED_DEPTHS.transform_keys(&:to_s),
  "formula_count" => formulae.length,
  "formulae" => formulae.sort_by { |formula| [formula.fetch("depth") || 99, formula.fetch("name")] }
}

output = Pathname.new(options[:output])
output.dirname.mkpath
output.write(JSON.pretty_generate(data) + "\n")

pending = formulae.count { |formula| formula.fetch("openssl_dependency") == "openssl@3" }
done = formulae.count { |formula| formula.fetch("openssl_dependency") == "openssl@4" }
puts "Wrote #{output} (#{formulae.length} formulae: #{pending} pending, #{done} done)"
